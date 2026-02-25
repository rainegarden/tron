package tabs

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	Path        string
	DisplayName string
	Dirty       bool
	Index       int
}

type TabBar struct {
	tabs         []*Tab
	activeIndex  int
	width        int
	maxTabWidth  int
	scrollOffset int
	height       int
}

func New() *TabBar {
	return &TabBar{
		tabs:         make([]*Tab, 0),
		activeIndex:  -1,
		width:        80,
		maxTabWidth:  30,
		scrollOffset: 0,
		height:       1,
	}
}

func (t *TabBar) AddTab(path string) int {
	displayName := filepath.Base(path)
	tab := &Tab{
		Path:        path,
		DisplayName: displayName,
		Dirty:       false,
		Index:       len(t.tabs),
	}
	t.tabs = append(t.tabs, tab)
	if t.activeIndex < 0 {
		t.activeIndex = 0
	}
	return tab.Index
}

func (t *TabBar) CloseTab(index int) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs = append(t.tabs[:index], t.tabs[index+1:]...)
	for i := range t.tabs {
		t.tabs[i].Index = i
	}
	if t.activeIndex >= len(t.tabs) {
		t.activeIndex = len(t.tabs) - 1
	}
	if t.activeIndex < 0 {
		t.activeIndex = 0
	}
	t.adjustScrollOffset()
}

func (t *TabBar) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.activeIndex = index
		t.ensureActiveVisible()
	}
}

func (t *TabBar) GetActive() *Tab {
	if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
		return t.tabs[t.activeIndex]
	}
	return nil
}

func (t *TabBar) MarkDirty(index int, dirty bool) {
	if index >= 0 && index < len(t.tabs) {
		t.tabs[index].Dirty = dirty
	}
}

func (t *TabBar) FindTab(path string) int {
	for i, tab := range t.tabs {
		if tab.Path == path {
			return i
		}
	}
	return -1
}

func (t *TabBar) SetSize(w, h int) {
	t.width = w
	t.height = h
	t.adjustScrollOffset()
}

func (t *TabBar) Init() tea.Cmd {
	return nil
}

func (t *TabBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return t.handleMouse(msg)
	case tea.KeyMsg:
		return t.handleKey(msg)
	}
	return t, nil
}

func (t *TabBar) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		return t, nil
	}
	if msg.Y != 0 {
		return t, nil
	}
	x := msg.X

	newBtnWidth := 3
	if x >= t.width-newBtnWidth {
		return t, t.newTabCmd()
	}

	for i := t.scrollOffset; i < len(t.tabs); i++ {
		tab := t.tabs[i]
		tabWidth := t.calculateTabWidth(tab)
		tabStart, tabEnd := t.getTabBounds(i)

		if x >= tabStart && x < tabEnd {
			closeBtnStart := tabEnd - 3
			if x >= closeBtnStart && x < tabEnd {
				return t, t.closeTabCmd(i, tab.Path)
			}
			return t, t.switchTabCmd(i, tab.Path)
		}
	}

	return t, nil
}

func (t *TabBar) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+tab":
		if len(t.tabs) == 0 {
			return t, nil
		}
		nextIndex := (t.activeIndex + 1) % len(t.tabs)
		t.SetActive(nextIndex)
		if tab := t.GetActive(); tab != nil {
			return t, t.switchTabCmd(nextIndex, tab.Path)
		}
	case "ctrl+w":
		if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
			tab := t.tabs[t.activeIndex]
			return t, t.closeTabCmd(t.activeIndex, tab.Path)
		}
	}
	return t, nil
}

func (t *TabBar) View() string {
	if t.width == 0 {
		return ""
	}

	var tabStrs []string
	for i := t.scrollOffset; i < len(t.tabs); i++ {
		tab := t.tabs[i]
		tabStr := t.renderTab(tab, i == t.activeIndex)
		tabStrs = append(tabStrs, tabStr)

		totalWidth := 0
		for _, s := range tabStrs {
			totalWidth += lipgloss.Width(s)
		}
		if totalWidth > t.width-3 {
			tabStrs = tabStrs[:len(tabStrs)-1]
			break
		}
	}

	newBtn := t.renderNewButton()
	remainingWidth := t.width
	for _, s := range tabStrs {
		remainingWidth -= lipgloss.Width(s)
	}
	if remainingWidth < 3 {
		remainingWidth = 3
	}

	tabBarStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1e1e2e"))
	var result string
	if len(tabStrs) > 0 {
		result = lipgloss.JoinHorizontal(lipgloss.Top, tabStrs...)
	}
	result = tabBarStyle.Render(result)

	padding := t.width - lipgloss.Width(result) - lipgloss.Width(newBtn)
	if padding > 0 {
		result += tabBarStyle.Render(strings.Repeat(" ", padding))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, result, newBtn)
}

func (t *TabBar) renderTab(tab *Tab, active bool) string {
	var style lipgloss.Style
	if active {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(lipgloss.Color("#cdd6f4")).
			Padding(0, 1)
	} else {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("#1e1e2e")).
			Foreground(lipgloss.Color("#6c7086")).
			Padding(0, 1)
	}

	displayName := tab.DisplayName
	if tab.Dirty {
		displayName = "* " + displayName
	}

	maxWidth := t.maxTabWidth - 4
	if maxWidth < 5 {
		maxWidth = 5
	}
	if len(displayName) > maxWidth {
		displayName = displayName[:maxWidth-1] + "…"
	}

	closeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f38ba8"))

	content := displayName + " " + closeStyle.Render("✕")

	return style.Render(content)
}

func (t *TabBar) renderNewButton() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#1e1e2e")).
		Foreground(lipgloss.Color("#89b4fa")).
		Padding(0, 1)

	return style.Render(" + ")
}

func (t *TabBar) calculateTabWidth(tab *Tab) int {
	displayName := tab.DisplayName
	if tab.Dirty {
		displayName = "* " + displayName
	}

	maxWidth := t.maxTabWidth - 4
	if maxWidth < 5 {
		maxWidth = 5
	}
	if len(displayName) > maxWidth {
		displayName = displayName[:maxWidth-1] + "…"
	}

	return len(displayName) + 6
}

func (t *TabBar) getTabBounds(index int) (int, int) {
	start := 0
	for i := t.scrollOffset; i < index; i++ {
		start += t.calculateTabWidth(t.tabs[i])
	}
	end := start + t.calculateTabWidth(t.tabs[index])
	return start, end
}

func (t *TabBar) adjustScrollOffset() {
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
	if t.scrollOffset > len(t.tabs) {
		t.scrollOffset = len(t.tabs) - 1
		if t.scrollOffset < 0 {
			t.scrollOffset = 0
		}
	}
}

func (t *TabBar) ensureActiveVisible() {
	if t.activeIndex < 0 {
		return
	}

	totalWidth := 0
	for i := t.scrollOffset; i <= t.activeIndex; i++ {
		totalWidth += t.calculateTabWidth(t.tabs[i])
	}

	availableWidth := t.width - 3
	for totalWidth > availableWidth && t.scrollOffset < t.activeIndex {
		totalWidth -= t.calculateTabWidth(t.tabs[t.scrollOffset])
		t.scrollOffset++
	}
}

func (t *TabBar) switchTabCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		return TabSwitchedMsg{Index: index, FilePath: path}
	}
}

func (t *TabBar) closeTabCmd(index int, path string) tea.Cmd {
	return func() tea.Msg {
		return TabClosedMsg{Index: index, FilePath: path}
	}
}

func (t *TabBar) newTabCmd() tea.Cmd {
	return func() tea.Msg {
		return NewTabMsg{}
	}
}

func (t *TabBar) TabCount() int {
	return len(t.tabs)
}

func (t *TabBar) GetTabs() []*Tab {
	return t.tabs
}

func (t *TabBar) GetTab(index int) *Tab {
	if index >= 0 && index < len(t.tabs) {
		return t.tabs[index]
	}
	return nil
}

func (t *TabBar) NextTab() {
	if len(t.tabs) == 0 {
		return
	}
	t.activeIndex = (t.activeIndex + 1) % len(t.tabs)
	t.ensureActiveVisible()
}

func (t *TabBar) PrevTab() {
	if len(t.tabs) == 0 {
		return
	}
	t.activeIndex--
	if t.activeIndex < 0 {
		t.activeIndex = len(t.tabs) - 1
	}
	t.ensureActiveVisible()
}

func (t *TabBar) UpdateTabPath(index int, newPath string) {
	if index >= 0 && index < len(t.tabs) {
		t.tabs[index].Path = newPath
		t.tabs[index].DisplayName = filepath.Base(newPath)
	}
}
