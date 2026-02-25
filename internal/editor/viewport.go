package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Viewport struct {
	Y      int
	X      int
	Height int
	Width  int
}

func NewViewport() *Viewport {
	return &Viewport{
		Y:      0,
		X:      0,
		Height: 24,
		Width:  80,
	}
}

func (v *Viewport) VisibleLineRange() (start, end int) {
	return v.Y, v.Y + v.Height
}

func (v *Viewport) VisibleColumnRange() (start, end int) {
	return v.X, v.X + v.Width
}

func (v *Viewport) IsLineVisible(line int) bool {
	return line >= v.Y && line < v.Y+v.Height
}

func (v *Viewport) IsColumnVisible(col int) bool {
	return col >= v.X && col < v.X+v.Width
}

func (v *Viewport) ScrollToLine(line int) {
	if line < v.Y {
		v.Y = line
	} else if line >= v.Y+v.Height {
		v.Y = line - v.Height + 1
	}
}

func (v *Viewport) ScrollToColumn(col int) {
	if col < v.X {
		v.X = col
	} else if col >= v.X+v.Width {
		v.X = col - v.Width + 1
	}
}

func (v *Viewport) ScrollUp() {
	if v.Y > 0 {
		v.Y--
	}
}

func (v *Viewport) ScrollDown(maxLines int) {
	if v.Y+v.Height < maxLines {
		v.Y++
	}
}

func (v *Viewport) ScrollLeft() {
	if v.X > 0 {
		v.X--
	}
}

func (v *Viewport) ScrollRight(maxCol int) {
	if v.X+v.Width < maxCol {
		v.X++
	}
}

func (v *Viewport) EnsureCursorVisible(cursor Position, lineLength int) {
	v.ScrollToLine(cursor.Line)
	v.ScrollToColumn(cursor.Column)
}

func (v *Viewport) HandleMouse(msg tea.MouseMsg, buffer Buffer) {
	switch msg.Type {
	case tea.MouseWheelUp:
		v.ScrollUp()
	case tea.MouseWheelDown:
		v.ScrollDown(buffer.LineCount())
	}
}
