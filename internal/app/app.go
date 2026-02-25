package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tron/internal/filetree"
	"tron/internal/runconfig"
	"tron/internal/tabs"
	"tron/internal/terminal"
	"tron/pkg/layout"
)

type headerPanel struct {
	tabs   *tabs.TabBar
	runBar *runconfig.RunBar
	width  int
	height int
}

func newHeaderPanel(rootPath string) *headerPanel {
	return &headerPanel{
		tabs:   tabs.New(),
		runBar: runconfig.NewRunBar(rootPath),
		height: 1,
	}
}

func (h *headerPanel) Init() tea.Cmd {
	return nil
}

func (h *headerPanel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if cmd := h.tabs.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := h.runBar.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (h *headerPanel) View() string {
	if h.width == 0 {
		return ""
	}

	tabsView := h.tabs.View()
	runBarView := h.runBar.View()

	tabsWidth := lipgloss.Width(tabsView)
	runBarWidth := lipgloss.Width(runBarView)

	remaining := h.width - runBarWidth
	if remaining < 0 {
		remaining = 0
	}

	if tabsWidth > remaining {
		h.tabs.SetSize(remaining, h.height)
		tabsView = h.tabs.View()
	}

	headerStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1e1e2e"))
	spacer := h.width - lipgloss.Width(tabsView) - lipgloss.Width(runBarView)
	if spacer < 0 {
		spacer = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		tabsView,
		headerStyle.Render(makeSpacer(spacer)),
		runBarView,
	)
}

func (h *headerPanel) SetSize(w, h int) {
	h.width = w
	h.height = h
	h.tabs.SetSize(w, h)
	h.runBar.SetSize(w, h)
}

func makeSpacer(n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]byte, n)
	for i := range result {
		result[i] = ' '
	}
	return string(result)
}

type Model struct {
	Width    int
	Height   int
	Root     layout.Panel
	FileTree *filetree.FileTree
	Tabs     *tabs.TabBar
	RunBar   *runconfig.RunBar
	header   *headerPanel
	Terminal *terminal.Terminal
}

func New() Model {
	ft := filetree.New(".")
	editor := layout.NewPlaceholderPanel("Editor")
	term := terminal.New()
	header := newHeaderPanel(".")

	editorTerminalSplit := layout.NewVerticalSplit(editor, term, 0.7)
	editorTerminalSplit.SetMinSizes(5, 3)

	mainSplit := layout.NewHorizontalSplit(ft, editorTerminalSplit, 0.2)
	mainSplit.SetMinSizes(15, 30)

	rootSplit := layout.NewVerticalSplit(header, mainSplit, 0.05)
	rootSplit.SetMinSizes(1, 5)

	return Model{
		Root:     rootSplit,
		FileTree: ft,
		Tabs:     header.tabs,
		RunBar:   header.runBar,
		header:   header,
		Terminal: term,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case filetree.FileSelectedMsg:
	case tabs.TabSwitchedMsg:
		m.Tabs.SetActive(msg.Index)
	case tabs.TabClosedMsg:
		m.Tabs.CloseTab(msg.Index)
	case tabs.NewTabMsg:
	case runconfig.RunCommandMsg:
		return m, m.handleRunCommand(msg)
	}

	var cmd tea.Cmd
	if cmd = m.Root.Update(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m Model) handleRunCommand(msg runconfig.RunCommandMsg) tea.Cmd {
	if msg.Config == nil {
		return nil
	}
	cmdParts := []string{msg.Config.Command}
	cmdParts = append(cmdParts, msg.Config.Args...)
	cmdStr := strings.Join(cmdParts, " ")

	cwd := msg.Config.WorkingDir
	if cwd == "" {
		cwd = "."
	}

	m.Terminal.RunCommand(cmdStr, cwd)
	return func() tea.Msg {
		return RunCommandMsg{
			Command: msg.Config.Command,
			Args:    msg.Config.Args,
		}
	}
}

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.Root.View()
}
