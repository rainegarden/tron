package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"tron/pkg/layout"
)

type Model struct {
	Width  int
	Height int
	Root   layout.Panel
}

func New() Model {
	fileTree := layout.NewPlaceholderPanel("FileTree")
	editor := layout.NewPlaceholderPanel("Editor")
	terminal := layout.NewPlaceholderPanel("Terminal")
	tabs := layout.NewPlaceholderPanel("Tabs")

	editorTerminalSplit := layout.NewVerticalSplit(editor, terminal, 0.7)
	editorTerminalSplit.SetMinSizes(5, 3)

	mainSplit := layout.NewHorizontalSplit(fileTree, editorTerminalSplit, 0.2)
	mainSplit.SetMinSizes(15, 30)

	rootSplit := layout.NewVerticalSplit(tabs, mainSplit, 0.05)
	rootSplit.SetMinSizes(1, 5)

	return Model{
		Root: rootSplit,
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
	}

	var cmd tea.Cmd
	if cmd = m.Root.Update(msg); cmd != nil {
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.Root.View()
}
