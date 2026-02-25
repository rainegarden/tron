package lsp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

func fileToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return "file://" + abs
}

func uriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

func getLanguageID(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescriptreact"
	case ".jsx":
		return "javascriptreact"
	case ".rs":
		return "rust"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".h":
		return "c"
	case ".hpp":
		return "cpp"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".sh":
		return "shellscript"
	default:
		return "plaintext"
	}
}

func (c *Client) Initialize(rootPath string) error {
	params := &InitializeParams{
		ProcessID: 0,
		RootURI:   fileToURI(rootPath),
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Completion: CompletionClientCapabilities{
					CompletionItem: CompletionItemCapabilities{
						SnippetSupport: true,
					},
				},
				Definition: DefinitionClientCapabilities{
					LinkSupport: true,
				},
				PublishDiagnostics: PublishDiagnosticsClientCapabilities{
					RelatedInformation: true,
				},
			},
			Workspace: WorkspaceClientCapabilities{
				WorkspaceFolders: true,
			},
		},
		Trace: "off",
	}

	id, err := c.SendRequest("initialize", params)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	resp, err := c.WaitForResponse(id)
	if err != nil {
		return fmt.Errorf("failed to receive initialize response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize failed: %s", resp.Error.Message)
	}

	c.SetInitialized(true)

	if err := c.SendNotification("initialized", struct{}{}); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

func (c *Client) OpenDocument(path string, content string) error {
	if !c.IsInitialized() {
		return fmt.Errorf("client not initialized")
	}

	params := &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        fileToURI(path),
			LanguageID: getLanguageID(path),
			Version:    1,
			Text:       content,
		},
	}

	return c.SendNotification("textDocument/didOpen", params)
}

func (c *Client) DidChangeDocument(path string, content string, version int) error {
	if !c.IsInitialized() {
		return fmt.Errorf("client not initialized")
	}

	params := &DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     fileToURI(path),
			Version: version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: content,
			},
		},
	}

	return c.SendNotification("textDocument/didChange", params)
}

func (c *Client) GetCompletions(path string, line, col int) ([]CompletionItem, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := &CompletionParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: fileToURI(path),
			},
			Position: Position{
				Line:      line,
				Character: col,
			},
		},
	}

	id, err := c.SendRequest("textDocument/completion", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send completion request: %w", err)
	}

	resp, err := c.WaitForResponse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to receive completion response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("completion failed: %s", resp.Error.Message)
	}

	var items []CompletionItem

	switch result := resp.Result.(type) {
	case nil:
		return nil, nil
	case []interface{}:
		for _, item := range result {
			var ci CompletionItem
			data, err := json.Marshal(item)
			if err != nil {
				continue
			}
			if err := json.Unmarshal(data, &ci); err != nil {
				continue
			}
			items = append(items, ci)
		}
	case map[string]interface{}:
		if itemsRaw, ok := result["items"]; ok {
			if itemsArray, ok := itemsRaw.([]interface{}); ok {
				for _, item := range itemsArray {
					var ci CompletionItem
					data, err := json.Marshal(item)
					if err != nil {
						continue
					}
					if err := json.Unmarshal(data, &ci); err != nil {
						continue
					}
					items = append(items, ci)
				}
			}
		}
	}

	return items, nil
}

func (c *Client) GoToDefinition(path string, line, col int) (*Location, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	params := &TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{
			URI: fileToURI(path),
		},
		Position: Position{
			Line:      line,
			Character: col,
		},
	}

	id, err := c.SendRequest("textDocument/definition", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send definition request: %w", err)
	}

	resp, err := c.WaitForResponse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to receive definition response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("definition failed: %s", resp.Error.Message)
	}

	switch result := resp.Result.(type) {
	case nil:
		return nil, nil
	case map[string]interface{}:
		var loc Location
		data, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &loc); err != nil {
			return nil, err
		}
		return &loc, nil
	case []interface{}:
		if len(result) > 0 {
			var loc Location
			data, err := json.Marshal(result[0])
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(data, &loc); err != nil {
				return nil, err
			}
			return &loc, nil
		}
	}

	return nil, nil
}

func (c *Client) Shutdown() error {
	if !c.IsInitialized() {
		return nil
	}

	id, err := c.SendRequest("shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to send shutdown request: %w", err)
	}

	_, err = c.WaitForResponse(id)
	if err != nil {
		return fmt.Errorf("failed to receive shutdown response: %w", err)
	}

	c.SendNotification("exit", nil)

	return nil
}
