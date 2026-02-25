package filetree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileTree struct {
	RootPath      string
	Nodes         []*Node
	Expanded      map[string]bool
	SelectedIndex int
	ScrollOffset  int
	Width         int
	Height        int
	ShowHidden    bool
	focused       bool
	flattened     []*displayItem
	lastClickTime int64
	lastClickY    int
}

type displayItem struct {
	Node     *Node
	Depth    int
	Path     string
}

func New(rootPath string) *FileTree {
	ft := &FileTree{
		RootPath: rootPath,
		Expanded: make(map[string]bool),
		ShowHidden: false,
		focused:  true,
	}
	ft.Refresh()
	return ft
}

func (ft *FileTree) Refresh() {
	ft.Nodes = ft.readDir(ft.RootPath)
	ft.flattenNodes()
}

func (ft *FileTree) readDir(path string) []*Node {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var nodes []*Node
	for _, entry := range entries {
		name := entry.Name()
		if !ft.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(path, name)
		isDir := entry.IsDir()

		node := &Node{
			Name:     name,
			Path:     fullPath,
			IsDir:    isDir,
			Expanded: ft.Expanded[fullPath],
		}

		if isDir && node.Expanded {
			node.Children = ft.readDir(fullPath)
		}

		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return nodes[i].Name < nodes[j].Name
	})

	return nodes
}

func (ft *FileTree) flattenNodes() {
	ft.flattened = nil
	for _, node := range ft.Nodes {
		ft.flattenNode(node, 0)
	}
}

func (ft *FileTree) flattenNode(node *Node, depth int) {
	ft.flattened = append(ft.flattened, &displayItem{
		Node:  node,
		Depth: depth,
		Path:  node.Path,
	})

	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			ft.flattenNode(child, depth+1)
		}
	}
}

func (ft *FileTree) Expand(path string) {
	ft.Expanded[path] = true
	ft.Refresh()
}

func (ft *FileTree) Collapse(path string) {
	ft.Expanded[path] = false
	ft.Refresh()
}

func (ft *FileTree) Toggle(path string) {
	if ft.Expanded[path] {
		ft.Collapse(path)
	} else {
		ft.Expand(path)
	}
}

func (ft *FileTree) SelectedPath() string {
	if ft.SelectedIndex < 0 || ft.SelectedIndex >= len(ft.flattened) {
		return ""
	}
	return ft.flattened[ft.SelectedIndex].Path
}

func (ft *FileTree) SetSize(w, h int) {
	ft.Width = w
	ft.Height = h
}

func (ft *FileTree) Focus() {
	ft.focused = true
}

func (ft *FileTree) Blur() {
	ft.focused = false
}

func (ft *FileTree) Focused() bool {
	return ft.focused
}

func (ft *FileTree) Init() tea.Cmd {
	return nil
}

func (ft *FileTree) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return ft.handleKey(msg)
	case tea.MouseMsg:
		return ft.handleMouse(msg)
	case FileTreeRefreshMsg:
		ft.Refresh()
		return nil
	}
	return nil
}

func (ft *FileTree) handleKey(msg tea.KeyMsg) tea.Cmd {
	if !ft.focused {
		return nil
	}

	switch msg.Type {
	case tea.KeyUp:
		ft.moveSelection(-1)
	case tea.KeyDown:
		ft.moveSelection(1)
	case tea.KeyEnter, tea.KeyRight:
		return ft.activateSelected()
	case tea.KeyLeft:
		ft.collapseOrGoUp()
	}

	switch msg.String() {
	case "h":
		ft.collapseOrGoUp()
	case "l":
		return ft.activateSelected()
	}

	return nil
}

func (ft *FileTree) moveSelection(delta int) {
	if len(ft.flattened) == 0 {
		return
	}

	ft.SelectedIndex += delta
	if ft.SelectedIndex < 0 {
		ft.SelectedIndex = 0
	}
	if ft.SelectedIndex >= len(ft.flattened) {
		ft.SelectedIndex = len(ft.flattened) - 1
	}

	ft.ensureSelectedVisible()
}

func (ft *FileTree) ensureSelectedVisible() {
	if ft.Height <= 0 {
		return
	}

	visibleStart := ft.ScrollOffset
	visibleEnd := ft.ScrollOffset + ft.Height

	if ft.SelectedIndex < visibleStart {
		ft.ScrollOffset = ft.SelectedIndex
	} else if ft.SelectedIndex >= visibleEnd {
		ft.ScrollOffset = ft.SelectedIndex - ft.Height + 1
	}
}

func (ft *FileTree) activateSelected() tea.Cmd {
	if ft.SelectedIndex < 0 || ft.SelectedIndex >= len(ft.flattened) {
		return nil
	}

	item := ft.flattened[ft.SelectedIndex]
	if item.Node.IsDir {
		ft.Toggle(item.Path)
		return nil
	}
	return func() tea.Msg {
		return FileSelectedMsg{Path: item.Path, IsDir: false}
	}
}

func (ft *FileTree) collapseOrGoUp() {
	if ft.SelectedIndex < 0 || ft.SelectedIndex >= len(ft.flattened) {
		return
	}

	item := ft.flattened[ft.SelectedIndex]
	if item.Node.IsDir && item.Node.Expanded {
		ft.Collapse(item.Path)
	}
}

func (ft *FileTree) handleMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg.Type {
	case tea.MouseLeft:
		localY := msg.Y
		idx := localY + ft.ScrollOffset
		if idx >= 0 && idx < len(ft.flattened) {
			now := time.Now().UnixMilli()
			if ft.lastClickY == localY && now-ft.lastClickTime < 500 {
				ft.SelectedIndex = idx
				return ft.activateSelected()
			}
			ft.lastClickTime = now
			ft.lastClickY = localY
			ft.SelectedIndex = idx
		}
	case tea.MouseWheelUp:
		if ft.ScrollOffset > 0 {
			ft.ScrollOffset--
		}
		if ft.SelectedIndex > 0 && ft.SelectedIndex >= ft.ScrollOffset+ft.Height {
			ft.SelectedIndex--
		}
	case tea.MouseWheelDown:
		if ft.ScrollOffset < len(ft.flattened)-ft.Height {
			ft.ScrollOffset++
		}
		if ft.SelectedIndex < len(ft.flattened)-1 && ft.SelectedIndex < ft.ScrollOffset {
			ft.SelectedIndex++
		}
	}
	return nil
}

func (ft *FileTree) View() string {
	if ft.Width == 0 || ft.Height == 0 {
		return ""
	}

	var lines []string
	endIdx := ft.ScrollOffset + ft.Height
	if endIdx > len(ft.flattened) {
		endIdx = len(ft.flattened)
	}

	for i := ft.ScrollOffset; i < endIdx; i++ {
		item := ft.flattened[i]
		line := ft.renderItem(item, i == ft.SelectedIndex)
		visualWidth := lipgloss.Width(line)
		if visualWidth > ft.Width {
			line = line[:ft.Width]
		} else if visualWidth < ft.Width {
			line += strings.Repeat(" ", ft.Width-visualWidth)
		}
		lines = append(lines, line)
	}

	for len(lines) < ft.Height {
		lines = append(lines, strings.Repeat(" ", ft.Width))
	}

	return strings.Join(lines, "\n")
}

func (ft *FileTree) renderItem(item *displayItem, selected bool) string {
	var sb strings.Builder

	indent := strings.Repeat("  ", item.Depth)
	sb.WriteString(indent)

	if item.Node.IsDir {
		if item.Node.Expanded {
			sb.WriteString("▾ ")
		} else {
			sb.WriteString("▸ ")
		}
	} else {
		sb.WriteString(ft.fileIcon(item.Node.Name))
		sb.WriteString(" ")
	}

	sb.WriteString(item.Node.Name)

	result := sb.String()

	if selected && ft.focused {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#4a9eff"))
		return style.Render(result)
	} else if selected {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("#333333"))
		return style.Render(result)
	}

	if item.Node.IsDir {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#4a9eff"))
		return style.Render(result)
	}

	return result
}

func (ft *FileTree) fileIcon(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return ""
	case ".js", ".jsx":
		return ""
	case ".ts", ".tsx":
		return ""
	case ".py":
		return ""
	case ".rs":
		return ""
	case ".rb":
		return ""
	case ".java":
		return ""
	case ".c", ".h":
		return ""
	case ".cpp", ".hpp":
		return ""
	case ".md":
		return ""
	case ".json":
		return ""
	case ".yaml", ".yml":
		return ""
	case ".toml":
		return ""
	case ".sh":
		return ""
	case ".txt":
		return ""
	case ".css":
		return ""
	case ".html":
		return ""
	case ".sql":
		return ""
	case ".png", ".jpg", ".jpeg", ".gif", ".svg":
		return ""
	case ".zip", ".tar", ".gz":
		return ""
	default:
		return ""
	}
}

func (ft *FileTree) ToggleHidden() {
	ft.ShowHidden = !ft.ShowHidden
	ft.Refresh()
}
