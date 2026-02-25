package app

type FileOpenMsg struct {
	Path string
}

type BufferChangeMsg struct {
	BufferID string
	Content  string
}

type TabSwitchMsg struct {
	TabID string
}

type RunCommandMsg struct {
	Command string
	Args    []string
}

type CommandCompleteMsg struct {
	Command string
	Output  string
	Error   error
}

type FocusChangeMsg struct {
	Component string
}

type FileTreeSelectMsg struct {
	Path  string
	IsDir bool
}

type EditorSaveMsg struct {
	Path    string
	Content string
}

type TerminalOutputMsg struct {
	Output string
}

type RunConfigSelectMsg struct {
	ConfigName string
}

type QuitMsg struct{}
