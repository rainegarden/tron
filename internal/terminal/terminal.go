package terminal

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Terminal struct {
	Lines       []string
	Command     string
	Cwd         string
	Cmd         *exec.Cmd
	Width       int
	Height      int
	ScrollPos   int
	AutoScroll  bool
	Running     bool
	ExitCode    int
	ExitError   error
	mu          sync.Mutex
	outputQueue []string
}

func New() *Terminal {
	return &Terminal{
		Lines:      make([]string, 0),
		AutoScroll: true,
		ExitCode:   -1,
	}
}

func (t *Terminal) RunCommand(cmdStr string, cwd string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Running {
		t.stopLocked()
	}

	t.Command = cmdStr
	t.Cwd = cwd
	t.Running = true
	t.ExitCode = -1
	t.ExitError = nil
	t.Lines = append(t.Lines, "")
	t.Lines = append(t.Lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Render("$ "+cmdStr))

	t.Cmd = exec.Command("sh", "-c", cmdStr)
	t.Cmd.Dir = cwd

	stdout, err := t.Cmd.StdoutPipe()
	if err != nil {
		t.Running = false
		return err
	}
	stderr, err := t.Cmd.StderrPipe()
	if err != nil {
		t.Running = false
		return err
	}

	if err := t.Cmd.Start(); err != nil {
		t.Running = false
		return err
	}

	go t.readOutput(stdout, false)
	go t.readOutput(stderr, true)

	go t.waitProcess()

	return nil
}

func (t *Terminal) readOutput(r io.Reader, isStderr bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = StripANSI(line)
		if isStderr {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")).Render(line)
		}
		t.mu.Lock()
		t.Lines = append(t.Lines, line)
		if t.AutoScroll {
			t.ScrollPos = len(t.Lines) - 1
		}
		t.mu.Unlock()
	}
}

func (t *Terminal) waitProcess() {
	err := t.Cmd.Wait()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Running = false
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				t.ExitCode = status.ExitStatus()
			} else {
				t.ExitCode = 1
			}
		} else {
			t.ExitCode = 1
		}
		t.ExitError = err
	} else {
		t.ExitCode = 0
	}
}

func (t *Terminal) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopLocked()
}

func (t *Terminal) stopLocked() {
	if t.Cmd != nil && t.Cmd.Process != nil {
		t.Cmd.Process.Kill()
		t.Running = false
		t.Lines = append(t.Lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("^C"))
	}
}

func (t *Terminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Lines = make([]string, 0)
	t.ScrollPos = 0
}

func (t *Terminal) ScrollUp() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AutoScroll = false
	if t.ScrollPos > 0 {
		t.ScrollPos--
	}
}

func (t *Terminal) ScrollDown() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ScrollPos < len(t.Lines)-1 {
		t.ScrollPos++
		if t.ScrollPos >= len(t.Lines)-1 {
			t.AutoScroll = true
		}
	}
}

func (t *Terminal) ScrollToBottom() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ScrollPos = len(t.Lines) - 1
	t.AutoScroll = true
}

func (t *Terminal) SetSize(w, h int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Width = w
	t.Height = h
}

func (t *Terminal) Init() tea.Cmd {
	return nil
}

func (t *Terminal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKey(msg)
	case tea.MouseMsg:
		return t.handleMouse(msg)
	case CommandStartedMsg:
		t.RunCommand(msg.Command, msg.Cwd)
		return t, nil
	}
	return t, nil
}

func (t *Terminal) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		t.ScrollUp()
	case tea.KeyDown:
		t.ScrollDown()
	case tea.KeyPgUp:
		for i := 0; i < t.Height-1 && t.ScrollPos > 0; i++ {
			t.ScrollUp()
		}
	case tea.KeyPgDown:
		for i := 0; i < t.Height-1 && t.ScrollPos < len(t.Lines)-1; i++ {
			t.ScrollDown()
		}
	default:
		switch msg.String() {
		case "ctrl+c":
			t.Stop()
		case "ctrl+l":
			t.Clear()
		}
	}
	return t, nil
}

func (t *Terminal) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseWheelUp:
		t.ScrollUp()
	case tea.MouseWheelDown:
		t.ScrollDown()
	}
	return t, nil
}

func (t *Terminal) View() string {
	if t.Width == 0 || t.Height == 0 {
		return ""
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	contentHeight := t.Height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	var visibleLines []string
	start := t.ScrollPos - contentHeight + 1
	if start < 0 {
		start = 0
	}
	end := start + contentHeight
	if end > len(t.Lines) {
		end = len(t.Lines)
	}
	if end > start {
		visibleLines = t.Lines[start:end]
	}

	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, "")
	}

	for i, line := range visibleLines {
		if len(line) > t.Width-2 {
			visibleLines[i] = line[:t.Width-2]
		}
	}

	content := strings.Join(visibleLines, "\n")

	scrollbar := t.renderScrollbar(start, len(t.Lines), contentHeight)

	statusBar := t.renderStatusBar()

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, content, scrollbar)

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, statusBar)
}

func (t *Terminal) renderScrollbar(start, total, height int) string {
	if total <= height {
		return lipgloss.NewStyle().
			Width(1).
			Height(height).
			Background(lipgloss.Color("#1e1e2e")).
			Render(" ")
	}

	trackStyle := lipgloss.NewStyle().Background(lipgloss.Color("#313244"))
	thumbStyle := lipgloss.NewStyle().Background(lipgloss.Color("#6c7086"))

	thumbHeight := max(1, height*height/total)
	thumbPos := start * height / total
	if thumbPos+thumbHeight > height {
		thumbPos = height - thumbHeight
	}

	var sb strings.Builder
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			sb.WriteString(thumbStyle.Render(" "))
		} else {
			sb.WriteString(trackStyle.Render(" "))
		}
		if i < height-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (t *Terminal) renderStatusBar() string {
	statusWidth := t.Width
	if statusWidth < 10 {
		statusWidth = 10
	}

	var status string
	if t.Running {
		spinner := "⠋"
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9e2af")).
			Render(spinner+" Running: "+t.Command)
	} else if t.ExitCode >= 0 {
		if t.ExitCode == 0 {
			status = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6e3a1")).
				Render("✓ Exit code: 0")
		} else {
			status = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f38ba8")).
				Render("✗ Exit code: "+string(rune('0'+t.ExitCode)))
		}
	} else {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")).
			Render("Ready")
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#313244")).
		Width(statusWidth)

	return style.Render(status)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
