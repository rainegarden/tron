package filetree

type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
	Expanded bool
}

type FileSelectedMsg struct {
	Path  string
	IsDir bool
}

type FileTreeRefreshMsg struct{}
