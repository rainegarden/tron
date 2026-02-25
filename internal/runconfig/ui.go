package runconfig

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RunBar struct {
	manager      *ConfigManager
	width        int
	height       int
	dropdownOpen bool
	focused      bool
}

func NewRunBar(rootPath string) *RunBar {
	return &RunBar{
		manager:      NewConfigManager(rootPath),
		dropdownOpen: false,
		focused:      false,
		height:       1,
	}
}

func (r *RunBar) Init() tea.Cmd {
	return nil
}

func (r *RunBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return r.handleKey(msg)
	case tea.MouseMsg:
		return r.handleMouse(msg)
	case ConfigSelectedMsg:
		r.manager.Select(msg.Index)
		r.dropdownOpen = false
		return r, nil
	case ShowDropdownMsg:
		r.dropdownOpen = true
		return r, nil
	case HideDropdownMsg:
		r.dropdownOpen = false
		return r, nil
	}
	return r, nil
}

func (r *RunBar) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f5", "ctrl+r":
		if config := r.manager.GetSelected(); config != nil {
			return r, r.runCommand(config)
		}
	case "up":
		if r.dropdownOpen && len(r.manager.Configs) > 0 {
			newIdx := r.manager.SelectedIndex - 1
			if newIdx < 0 {
				newIdx = len(r.manager.Configs) - 1
			}
			r.manager.Select(newIdx)
		}
	case "down":
		if r.dropdownOpen && len(r.manager.Configs) > 0 {
			newIdx := r.manager.SelectedIndex + 1
			if newIdx >= len(r.manager.Configs) {
				newIdx = 0
			}
			r.manager.Select(newIdx)
		}
	case "enter":
		if r.dropdownOpen {
			r.dropdownOpen = false
			if config := r.manager.GetSelected(); config != nil {
				return r, r.runCommand(config)
			}
		}
	case "esc":
		if r.dropdownOpen {
			r.dropdownOpen = false
			return r, nil
		}
	}
	return r, nil
}

func (r *RunBar) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		return r, nil
	}

	runBtnWidth := lipgloss.Width(r.renderRunButton())
	dropdownWidth := lipgloss.Width(r.renderDropdownButton())
	editBtnWidth := lipgloss.Width(r.renderEditButton())

	x := msg.X

	if x < runBtnWidth {
		if config := r.manager.GetSelected(); config != nil {
			return r, r.runCommand(config)
		}
		return r, nil
	}

	x -= runBtnWidth
	if x < dropdownWidth {
		r.dropdownOpen = !r.dropdownOpen
		return r, nil
	}

	if r.dropdownOpen {
		dropdownX := runBtnWidth + dropdownWidth
		dropdownItemHeight := 1
		for i, cfg := range r.manager.Configs {
			itemY := 1
			if msg.Y == itemY && msg.X >= dropdownX && msg.X < dropdownX+lipgloss.Width(r.renderDropdownItem(cfg, i == r.manager.SelectedIndex)) {
				r.manager.Select(i)
				r.dropdownOpen = false
				return r, r.selectConfigCmd(i, cfg.Name)
			}
		}
		_ = dropdownItemHeight
	}

	x -= dropdownWidth
	if x < editBtnWidth {
		return r, r.editConfigCmd(r.manager.SelectedIndex)
	}

	return r, nil
}

func (r *RunBar) View() string {
	if r.width == 0 {
		return ""
	}

	runBtn := r.renderRunButton()
	dropdownBtn := r.renderDropdownButton()
	editBtn := r.renderEditButton()

	bar := lipgloss.JoinHorizontal(lipgloss.Top, runBtn, dropdownBtn, editBtn)

	if r.dropdownOpen {
		dropdown := r.renderDropdown()
		return lipgloss.JoinVertical(lipgloss.Left, bar, dropdown)
	}

	return bar
}

func (r *RunBar) renderRunButton() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#a6e3a1")).
		Foreground(lipgloss.Color("#1e1e2e")).
		Padding(0, 1).
		Bold(true)

	return style.Render(" ▶ Run ")
}

func (r *RunBar) renderDropdownButton() string {
	config := r.manager.GetSelected()
	name := "No Config"
	if config != nil {
		name = config.Name
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#313244")).
		Foreground(lipgloss.Color("#cdd6f4")).
		Padding(0, 1)

	arrow := " ▼"
	if r.dropdownOpen {
		arrow = " ▲"
	}

	return style.Render(" " + name + arrow + " ")
}

func (r *RunBar) renderEditButton() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#45475a")).
		Foreground(lipgloss.Color("#cdd6f4")).
		Padding(0, 1)

	return style.Render(" ⚙ ")
}

func (r *RunBar) renderDropdown() string {
	if len(r.manager.Configs) == 0 {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(lipgloss.Color("#6c7086")).
			Padding(0, 1)
		return style.Render(" No configs available ")
	}

	items := make([]string, 0, len(r.manager.Configs))
	for i, cfg := range r.manager.Configs {
		items = append(items, r.renderDropdownItem(cfg, i == r.manager.SelectedIndex))
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#313244"))

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

func (r *RunBar) renderDropdownItem(cfg *RunConfig, selected bool) string {
	var style lipgloss.Style
	if selected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("#89b4fa")).
			Foreground(lipgloss.Color("#1e1e2e")).
			Padding(0, 1).
			Width(20)
	} else {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(lipgloss.Color("#cdd6f4")).
			Padding(0, 1).
			Width(20)
	}

	return style.Render(" " + cfg.Name)
}

func (r *RunBar) SetSize(w, h int) {
	r.width = w
	r.height = h
}

func (r *RunBar) runCommand(config *RunConfig) tea.Cmd {
	return func() tea.Msg {
		return RunCommandMsg{Config: config}
	}
}

func (r *RunBar) selectConfigCmd(index int, name string) tea.Cmd {
	return func() tea.Msg {
		return ConfigSelectedMsg{Index: index, Name: name}
	}
}

func (r *RunBar) editConfigCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return EditConfigMsg{Index: index}
	}
}

func (r *RunBar) GetManager() *ConfigManager {
	return r.manager
}

func (r *RunBar) SetFocused(focused bool) {
	r.focused = focused
}

func (r *RunBar) IsDropdownOpen() bool {
	return r.dropdownOpen
}

func (r *RunBar) CloseDropdown() {
	r.dropdownOpen = false
}
