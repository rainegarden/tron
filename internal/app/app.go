package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Width  int
	Height int

	FileTree  FileTreePlaceholder
	Editor    EditorPlaceholder
	Tabs      TabsPlaceholder
	Terminal  TerminalPlaceholder
	RunConfig RunConfigPlaceholder
}

type FileTreePlaceholder struct {
	Width  int
	Height int
	Files  []string
}

type EditorPlaceholder struct {
	Width   int
	Height  int
	Content string
	Path    string
}

type TabsPlaceholder struct {
	Width   int
	Height  int
	Tabs    []string
	Active  int
}

type TerminalPlaceholder struct {
	Width   int
	Height  int
	Content string
}

type RunConfigPlaceholder struct {
	Width   int
	Height  int
	Configs []string
	Active  string
}

func New() Model {
	return Model{
		FileTree: FileTreePlaceholder{
			Files: []string{},
		},
		Editor: EditorPlaceholder{
			Content: "",
		},
		Tabs: TabsPlaceholder{
			Tabs:   []string{},
			Active: 0,
		},
		Terminal: TerminalPlaceholder{
			Content: "",
		},
		RunConfig: RunConfigPlaceholder{
			Configs: []string{},
		},
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
	case tea.MouseMsg:
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.FileTree.Width = msg.Width / 4
		m.FileTree.Height = msg.Height - 2
		m.Editor.Width = msg.Width * 3 / 4
		m.Editor.Height = msg.Height - 4
		m.Tabs.Width = msg.Width
		m.Tabs.Height = 1
		m.Terminal.Width = msg.Width * 3 / 4
		m.Terminal.Height = msg.Height / 3
		m.RunConfig.Width = msg.Width / 4
		m.RunConfig.Height = msg.Height / 3
	case FileOpenMsg, BufferChangeMsg, TabSwitchMsg, RunCommandMsg,
		CommandCompleteMsg, FocusChangeMsg, FileTreeSelectMsg,
		EditorSaveMsg, TerminalOutputMsg, RunConfigSelectMsg, QuitMsg:
	}
	return m, nil
}

func (m Model) View() string {
	return fmt.Sprintf(
		"TRON IDE - %dx%d\n\n[FileTree]  [Editor]\n            [Terminal]\n[Tabs]\n[RunConfig]\n\nPress Ctrl+C to quit",
		m.Width, m.Height,
	)
}
