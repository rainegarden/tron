package tabs

type TabSwitchedMsg struct {
	Index    int
	FilePath string
}

type TabClosedMsg struct {
	Index    int
	FilePath string
}

type NewTabMsg struct{}

type TabAddedMsg struct {
	Index    int
	FilePath string
}
