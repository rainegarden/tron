package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Panel interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
	SetSize(w, h int)
}

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

type Split struct {
	Direction Direction
	First     Panel
	Second    Panel

	position    int
	ratio       float64
	width       int
	height      int
	minFirst    int
	minSecond   int
	dragging    bool
	dragOffset  int
	dividerSize int
}

func NewHorizontalSplit(left, right Panel, initialRatio float64) *Split {
	return &Split{
		Direction:   Horizontal,
		First:       left,
		Second:      right,
		ratio:       initialRatio,
		minFirst:    10,
		minSecond:   10,
		dividerSize: 1,
	}
}

func NewVerticalSplit(top, bottom Panel, initialRatio float64) *Split {
	return &Split{
		Direction:   Vertical,
		First:       top,
		Second:      bottom,
		ratio:       initialRatio,
		minFirst:    3,
		minSecond:   3,
		dividerSize: 1,
	}
}

func (s *Split) SetSize(w, h int) {
	s.width = w
	s.height = h
	s.recalculateSizes()
}

func (s *Split) recalculateSizes() {
	if s.width == 0 || s.height == 0 {
		return
	}

	if s.Direction == Horizontal {
		s.position = int(float64(s.width) * s.ratio)
		if s.position < s.minFirst {
			s.position = s.minFirst
		}
		if s.position > s.width-s.minSecond-s.dividerSize {
			s.position = s.width - s.minSecond - s.dividerSize
		}
		firstWidth := s.position
		secondWidth := s.width - s.position - s.dividerSize
		s.First.SetSize(firstWidth, s.height)
		s.Second.SetSize(secondWidth, s.height)
	} else {
		s.position = int(float64(s.height) * s.ratio)
		if s.position < s.minFirst {
			s.position = s.minFirst
		}
		if s.position > s.height-s.minSecond-s.dividerSize {
			s.position = s.height - s.minSecond - s.dividerSize
		}
		firstHeight := s.position
		secondHeight := s.height - s.position - s.dividerSize
		s.First.SetSize(s.width, firstHeight)
		s.Second.SetSize(s.width, secondHeight)
	}
}

func (s *Split) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return s.handleMouse(msg)
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)
		return nil
	}

	var cmds []tea.Cmd
	if cmd := s.First.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := s.Second.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *Split) handleMouse(msg tea.MouseMsg) tea.Cmd {
	if s.width == 0 || s.height == 0 {
		return nil
	}

	var dividerStart, dividerEnd int
	if s.Direction == Horizontal {
		dividerStart = s.position
		dividerEnd = s.position + s.dividerSize
	} else {
		dividerStart = s.position
		dividerEnd = s.position + s.dividerSize
	}

	isOverDivider := false
	if s.Direction == Horizontal {
		isOverDivider = msg.X >= dividerStart && msg.X < dividerEnd && msg.Y >= 0 && msg.Y < s.height
	} else {
		isOverDivider = msg.X >= 0 && msg.X < s.width && msg.Y >= dividerStart && msg.Y < dividerEnd
	}

	switch msg.Type {
	case tea.MouseLeft:
		if isOverDivider {
			s.dragging = true
			if s.Direction == Horizontal {
				s.dragOffset = msg.X - s.position
			} else {
				s.dragOffset = msg.Y - s.position
			}
			return nil
		}
	case tea.MouseRelease:
		s.dragging = false
		s.dragOffset = 0
		return nil
	case tea.MouseMotion:
		if s.dragging {
			var newPos int
			if s.Direction == Horizontal {
				newPos = msg.X - s.dragOffset
			} else {
				newPos = msg.Y - s.dragOffset
			}
			s.setPosition(newPos)
			return nil
		}
	}

	var cmds []tea.Cmd
	if cmd := s.First.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := s.Second.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *Split) setPosition(pos int) {
	maxPos := 0
	if s.Direction == Horizontal {
		maxPos = s.width - s.minSecond - s.dividerSize
	} else {
		maxPos = s.height - s.minSecond - s.dividerSize
	}

	if pos < s.minFirst {
		pos = s.minFirst
	}
	if pos > maxPos {
		pos = maxPos
	}

	s.position = pos

	if s.Direction == Horizontal {
		if s.width > 0 {
			s.ratio = float64(s.position) / float64(s.width)
		}
	} else {
		if s.height > 0 {
			s.ratio = float64(s.position) / float64(s.height)
		}
	}

	s.recalculateSizes()
}

func (s *Split) View() string {
	if s.width == 0 || s.height == 0 {
		return ""
	}

	dividerStyle := lipgloss.NewStyle()
	if s.dragging {
		dividerStyle = dividerStyle.Background(lipgloss.Color("62"))
	} else {
		dividerStyle = dividerStyle.Background(lipgloss.Color("238"))
	}

	var divider string
	if s.Direction == Horizontal {
		divider = dividerStyle.Width(s.dividerSize).Height(s.height).Render("")
	} else {
		divider = dividerStyle.Width(s.width).Height(s.dividerSize).Render("")
	}

	firstView := s.First.View()
	secondView := s.Second.View()

	if s.Direction == Horizontal {
		return lipgloss.JoinHorizontal(lipgloss.Top, firstView, divider, secondView)
	}
	return lipgloss.JoinVertical(lipgloss.Left, firstView, divider, secondView)
}

func (s *Split) SetMinSizes(minFirst, minSecond int) {
	s.minFirst = minFirst
	s.minSecond = minSecond
	s.recalculateSizes()
}

func (s *Split) IsDragging() bool {
	return s.dragging
}

type PlaceholderPanel struct {
	Title  string
	Width  int
	Height int
	Style  lipgloss.Style
}

func NewPlaceholderPanel(title string) *PlaceholderPanel {
	return &PlaceholderPanel{
		Title: title,
		Style: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")),
	}
}

func (p *PlaceholderPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (p *PlaceholderPanel) View() string {
	style := p.Style.Width(p.Width).Height(p.Height)
	content := style.Render(p.Title)
	return content
}

func (p *PlaceholderPanel) SetSize(w, h int) {
	p.Width = w
	p.Height = h
}
