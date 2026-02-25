package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type pendingRequest struct {
	response chan *Response
	err      chan error
}

type Client struct {
	cmd           string
	cmdArgs       []string
	process       *exec.Cmd
	stdin         io.WriteCloser
	stdout        *bufio.Reader
	stderr        io.ReadCloser
	requestID     atomic.Int32
	pending       map[int]*pendingRequest
	pendingMu     sync.RWMutex
	diagnostics   map[string][]Diagnostic
	diagnosticsMu sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	initialized   bool
	initMu        sync.RWMutex
}

func New(command string) *Client {
	return &Client{
		cmd:         command,
		cmdArgs:     []string{},
		pending:     make(map[int]*pendingRequest),
		diagnostics: make(map[string][]Diagnostic),
	}
}

func NewWithArgs(command string, args []string) *Client {
	return &Client{
		cmd:         command,
		cmdArgs:     args,
		pending:     make(map[int]*pendingRequest),
		diagnostics: make(map[string][]Diagnostic),
	}
}

func (c *Client) Start(rootPath string) error {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	c.process = exec.Command(c.cmd, c.cmdArgs...)
	c.process.Dir = absPath

	stdin, err := c.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdout)

	stderr, err := c.process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderr

	if err := c.process.Start(); err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.wg.Add(1)
	go c.readLoop()

	c.wg.Add(1)
	go c.stderrLoop()

	return nil
}

func (c *Client) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}

	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.process != nil && c.process.Process != nil {
		c.process.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() {
			done <- c.process.Wait()
		}()

		select {
		case err := <-done:
			c.wg.Wait()
			return err
		case <-c.ctx.Done():
			c.process.Process.Kill()
			c.wg.Wait()
			return c.process.Wait()
		}
	}

	c.wg.Wait()
	return nil
}

func (c *Client) SendRequest(method string, params interface{}) (int, error) {
	id := int(c.requestID.Add(1))

	req := &Request{
		JsonRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	c.pendingMu.Lock()
	c.pending[id] = &pendingRequest{
		response: make(chan *Response, 1),
		err:      make(chan error, 1),
	}
	c.pendingMu.Unlock()

	if err := WriteMessage(c.stdin, req); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return 0, fmt.Errorf("failed to send request: %w", err)
	}

	return id, nil
}

func (c *Client) SendNotification(method string, params interface{}) error {
	notif := &Notification{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	if err := WriteMessage(c.stdin, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (c *Client) WaitForResponse(id int) (*Response, error) {
	c.pendingMu.RLock()
	pending, ok := c.pending[id]
	c.pendingMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no pending request with id %d", id)
	}

	select {
	case resp := <-pending.response:
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return resp, nil
	case err := <-pending.err:
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, err
	case <-c.ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, c.ctx.Err()
	}
}

func (c *Client) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			data, err := ReadMessage(c.stdout)
			if err != nil {
				if c.ctx.Err() != nil {
					return
				}
				continue
			}

			var baseMsg struct {
				JsonRPC string     `json:"jsonrpc"`
				ID      *int       `json:"id"`
				Method  string     `json:"method,omitempty"`
				Error   *LSPError  `json:"error,omitempty"`
			}

			if err := json.Unmarshal(data, &baseMsg); err != nil {
				continue
			}

			if baseMsg.ID != nil {
				var resp Response
				if err := json.Unmarshal(data, &resp); err != nil {
					c.pendingMu.RLock()
					if pending, ok := c.pending[*baseMsg.ID]; ok {
						pending.err <- fmt.Errorf("failed to parse response: %w", err)
					}
					c.pendingMu.RUnlock()
					continue
				}

				c.pendingMu.RLock()
				if pending, ok := c.pending[resp.ID]; ok {
					pending.response <- &resp
				}
				c.pendingMu.RUnlock()
			} else if baseMsg.Method != "" {
				c.handleNotification(baseMsg.Method, data)
			}
		}
	}
}

func (c *Client) stderrLoop() {
	defer c.wg.Done()

	buf := make([]byte, 1024)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			n, err := c.stderr.Read(buf)
			if err != nil {
				return
			}
			_ = n
		}
	}
}

func (c *Client) handleNotification(method string, data []byte) {
	switch method {
	case "textDocument/publishDiagnostics":
		var params PublishDiagnosticsParams
		var notif struct {
			Params PublishDiagnosticsParams `json:"params"`
		}
		if err := json.Unmarshal(data, &notif); err != nil {
			return
		}
		params = notif.Params

		c.diagnosticsMu.Lock()
		c.diagnostics[params.URI] = params.Diagnostics
		c.diagnosticsMu.Unlock()
	}
}

func (c *Client) GetDiagnostics(uri string) []Diagnostic {
	c.diagnosticsMu.RLock()
	defer c.diagnosticsMu.RUnlock()
	return c.diagnostics[uri]
}

func (c *Client) ClearDiagnostics(uri string) {
	c.diagnosticsMu.Lock()
	defer c.diagnosticsMu.Unlock()
	delete(c.diagnostics, uri)
}

func (c *Client) IsInitialized() bool {
	c.initMu.RLock()
	defer c.initMu.RUnlock()
	return c.initialized
}

func (c *Client) SetInitialized(v bool) {
	c.initMu.Lock()
	defer c.initMu.Unlock()
	c.initialized = v
}
