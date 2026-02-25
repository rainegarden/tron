package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tron/internal/syntax"
)

type Editor struct {
	Buffer             Buffer
	Viewport           *Viewport
	Cursor             Position
	Selection          Selection
	Width              int
	Height             int
	CursorStyle        CursorStyle
	ShowLineNumbers    bool
	SelectionColor     string
	LineNumWidth       int
	ShowCursor         bool
	focused            bool
	anchor             Position
	selectionActive    bool
	fileExt            string
	highlightedContent string
	highlightSpans     []syntax.HighlightSpan
	theme              *syntax.Theme
	FilePath           string
	Dirty              bool
	originalContent    string
}

type EditorSavedMsg struct {
	Path string
}

type EditorDirtyMsg struct {
	Dirty bool
}

type EditorFocusMsg struct{}
type EditorBlurMsg struct{}

func New() *Editor {
	return &Editor{
		Buffer:          NewSimpleBuffer(),
		Viewport:        NewViewport(),
		Cursor:          Position{Line: 0, Column: 0},
		Selection:       Selection{},
		Width:           80,
		Height:          24,
		CursorStyle:     CursorBlock,
		ShowLineNumbers: true,
		SelectionColor:  "#334466",
		LineNumWidth:    4,
		ShowCursor:      true,
		focused:         true,
		theme:           syntax.GetTheme(),
	}
}

func NewWithBuffer(b Buffer) *Editor {
	e := New()
	e.Buffer = b
	return e
}

func NewWithContent(content string) *Editor {
	return NewWithBuffer(NewSimpleBufferWithContent(content))
}

func (e *Editor) SetSize(width, height int) {
	e.Width = width
	e.Height = height
	e.Viewport.Width = width - e.lineNumWidth()
	e.Viewport.Height = height
}

func (e *Editor) SetContent(content string) {
	e.Buffer.SetContent(content)
	e.Cursor = Position{Line: 0, Column: 0}
	e.Viewport.Y = 0
	e.Viewport.X = 0
	e.clearSelection()
	e.updateHighlighting()
}

func (e *Editor) SetFileExtension(ext string) {
	e.fileExt = ext
	e.updateHighlighting()
}

func (e *Editor) SetFilePath(path string) {
	e.fileExt = filepath.Ext(path)
	e.updateHighlighting()
}

func (e *Editor) updateHighlighting() {
	content := e.Buffer.Content()
	if content != e.highlightedContent {
		e.highlightedContent = content
		e.highlightSpans = syntax.Highlight(content, e.fileExt)
	}
}

func (e *Editor) Content() string {
	return e.Buffer.Content()
}

func (e *Editor) Focus() {
	e.focused = true
}

func (e *Editor) Blur() {
	e.focused = false
}

func (e *Editor) Focused() bool {
	return e.focused
}

func (e *Editor) Init() tea.Cmd {
	return nil
}

func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return e.handleKeyPress(msg)
	case tea.MouseMsg:
		return e.handleMouse(msg)
	case EditorFocusMsg:
		e.Focus()
		return e, nil
	case EditorBlurMsg:
		e.Blur()
		return e, nil
	}
	return e, nil
}

func (e *Editor) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !e.focused {
		return e, nil
	}

	switch msg.Type {
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			if e.hasSelection() {
				e.deleteSelection()
			}
			for _, r := range msg.Runes {
				e.Buffer.Insert(e.Cursor, string(r))
				e.Cursor.Column++
			}
			e.clearSelection()
			e.markDirty()
		}
	case tea.KeyEnter:
		if e.hasSelection() {
			e.deleteSelection()
		}
		e.Buffer.Insert(e.Cursor, "\n")
		e.Cursor.Line++
		e.Cursor.Column = 0
		e.clearSelection()
		e.markDirty()
	case tea.KeyBackspace:
		if e.hasSelection() {
			e.deleteSelection()
		} else {
			e.Buffer.DeleteChar(e.Cursor, false)
			if e.Cursor.Column > 0 {
				e.Cursor.Column--
			} else if e.Cursor.Line > 0 {
				e.Cursor.Line--
				e.Cursor.Column = e.Buffer.LineLength(e.Cursor.Line)
			}
		}
		e.clearSelection()
		e.markDirty()
	case tea.KeyDelete:
		if e.hasSelection() {
			e.deleteSelection()
		} else {
			e.Buffer.DeleteChar(e.Cursor, true)
		}
		e.clearSelection()
		e.markDirty()
	case tea.KeyLeft:
		e.moveCursor(-1, 0, msg.Modifiers)
	case tea.KeyRight:
		e.moveCursor(1, 0, msg.Modifiers)
	case tea.KeyUp:
		e.moveCursor(0, -1, msg.Modifiers)
	case tea.KeyDown:
		e.moveCursor(0, 1, msg.Modifiers)
	case tea.KeyHome:
		if msg.Modifiers == tea.ModAlt || msg.Modifiers == tea.ModCtrl {
			e.Cursor.Line = 0
			e.Cursor.Column = 0
		} else {
			e.Cursor.Column = 0
		}
		if msg.Modifiers == tea.ModShift {
			e.extendSelection()
		} else {
			e.clearSelection()
		}
	case tea.KeyEnd:
		if msg.Modifiers == tea.ModAlt || msg.Modifiers == tea.ModCtrl {
			e.Cursor.Line = e.Buffer.LineCount() - 1
			e.Cursor.Column = e.Buffer.LineLength(e.Cursor.Line)
		} else {
			e.Cursor.Column = e.Buffer.LineLength(e.Cursor.Line)
		}
		if msg.Modifiers == tea.ModShift {
			e.extendSelection()
		} else {
			e.clearSelection()
		}
	default:
		switch msg.String() {
		case "ctrl+a":
			e.selectAll()
		case "ctrl+c":
			e.copySelection()
		case "ctrl+v":
			e.paste()
			e.markDirty()
		case "ctrl+x":
			e.cutSelection()
			e.markDirty()
		case "ctrl+s":
			if e.FilePath != "" {
				if err := e.Save(); err == nil {
					return e, func() tea.Msg {
						return EditorSavedMsg{Path: e.FilePath}
					}
				}
			}
		}
	}

	e.ensureCursorValid()
	e.Viewport.EnsureCursorVisible(e.Cursor, e.Buffer.LineLength(e.Cursor.Line))
	e.updateHighlighting()
	return e, nil
}

func (e *Editor) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.MouseLeft:
		line := msg.Y - 1 + e.Viewport.Y
		col := msg.X - e.lineNumWidth() + e.Viewport.X
		if line >= 0 && line < e.Buffer.LineCount() {
			e.Cursor.Line = line
			e.Cursor.Column = max(0, min(col, e.Buffer.LineLength(line)))
		}
		e.clearSelection()
		e.selectionActive = true
		e.anchor = e.Cursor
	case tea.MouseRelease:
		e.selectionActive = false
	case tea.MouseMotion:
		if e.selectionActive {
			line := msg.Y - 1 + e.Viewport.Y
			col := msg.X - e.lineNumWidth() + e.Viewport.X
			if line >= 0 && line < e.Buffer.LineCount() {
				e.Cursor.Line = line
				e.Cursor.Column = max(0, min(col, e.Buffer.LineLength(line)))
				e.Selection.Start = e.anchor
				e.Selection.End = e.Cursor
			}
		}
	case tea.MouseWheelUp:
		e.Viewport.ScrollUp()
	case tea.MouseWheelDown:
		e.Viewport.ScrollDown(e.Buffer.LineCount())
	}
	return e, nil
}

func (e *Editor) moveCursor(dx, dy int, mods tea.KeyMod) {
	if mods == tea.ModShift {
		if !e.hasSelection() {
			e.Selection.Start = e.Cursor
		}
	}

	if dx != 0 {
		if dx < 0 {
			if e.Cursor.Column > 0 {
				e.Cursor.Column--
			} else if e.Cursor.Line > 0 {
				e.Cursor.Line--
				e.Cursor.Column = e.Buffer.LineLength(e.Cursor.Line)
			}
		} else {
			if e.Cursor.Column < e.Buffer.LineLength(e.Cursor.Line) {
				e.Cursor.Column++
			} else if e.Cursor.Line < e.Buffer.LineCount()-1 {
				e.Cursor.Line++
				e.Cursor.Column = 0
			}
		}
	}

	if dy != 0 {
		e.Cursor.Line += dy
		if e.Cursor.Line < 0 {
			e.Cursor.Line = 0
		} else if e.Cursor.Line >= e.Buffer.LineCount() {
			e.Cursor.Line = e.Buffer.LineCount() - 1
		}
		maxCol := e.Buffer.LineLength(e.Cursor.Line)
		if e.Cursor.Column > maxCol {
			e.Cursor.Column = maxCol
		}
	}

	if mods == tea.ModShift {
		e.Selection.End = e.Cursor
	} else {
		e.clearSelection()
	}
}

func (e *Editor) ensureCursorValid() {
	if e.Buffer.LineCount() == 0 {
		e.Cursor = Position{Line: 0, Column: 0}
		return
	}

	if e.Cursor.Line < 0 {
		e.Cursor.Line = 0
	} else if e.Cursor.Line >= e.Buffer.LineCount() {
		e.Cursor.Line = e.Buffer.LineCount() - 1
	}

	maxCol := e.Buffer.LineLength(e.Cursor.Line)
	if e.Cursor.Column < 0 {
		e.Cursor.Column = 0
	} else if e.Cursor.Column > maxCol {
		e.Cursor.Column = maxCol
	}
}

func (e *Editor) hasSelection() bool {
	return !e.Selection.IsEmpty()
}

func (e *Editor) clearSelection() {
	e.Selection = Selection{}
	e.selectionActive = false
}

func (e *Editor) extendSelection() {
	if !e.hasSelection() {
		e.Selection.Start = e.anchor
	}
	e.Selection.End = e.Cursor
}

func (e *Editor) selectAll() {
	e.Selection.Start = Position{Line: 0, Column: 0}
	lastLine := e.Buffer.LineCount() - 1
	e.Selection.End = Position{Line: lastLine, Column: e.Buffer.LineLength(lastLine)}
	e.Cursor = e.Selection.End
}

func (e *Editor) deleteSelection() {
	if !e.hasSelection() {
		return
	}
	norm := e.Selection.Normalized()
	e.Buffer.Delete(norm.Start, norm.End)
	e.Cursor = norm.Start
	e.clearSelection()
}

func (e *Editor) copySelection() {
	if !e.hasSelection() {
		return
	}
	norm := e.Selection.Normalized()
	text := e.Buffer.GetText(norm.Start, norm.End)
	_ = clipboard.WriteAll(text)
}

func (e *Editor) paste() {
	text, err := clipboard.ReadAll()
	if err != nil {
		return
	}
	if e.hasSelection() {
		e.deleteSelection()
	}
	e.Buffer.Insert(e.Cursor, text)
	e.moveCursorAfterInsert(text)
	e.clearSelection()
}

func (e *Editor) cutSelection() {
	if !e.hasSelection() {
		line := e.Buffer.Lines()[e.Cursor.Line]
		e.Selection.Start = Position{Line: e.Cursor.Line, Column: 0}
		e.Selection.End = Position{Line: e.Cursor.Line, Column: len(line)}
	}
	e.copySelection()
	e.deleteSelection()
}

func (e *Editor) LoadFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	e.FilePath = path
	e.SetContent(string(content))
	e.originalContent = string(content)
	e.Dirty = false
	e.SetFileExtension(path)
	return nil
}

func (e *Editor) Save() error {
	if e.FilePath == "" {
		return fmt.Errorf("no file path set")
	}
	content := e.Buffer.Content()
	err := os.WriteFile(e.FilePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	e.originalContent = content
	e.Dirty = false
	return nil
}

func (e *Editor) SaveAs(path string) error {
	e.FilePath = path
	return e.Save()
}

func (e *Editor) markDirty() {
	if !e.Dirty {
		e.Dirty = true
	}
}

func (e *Editor) IsDirty() bool {
	return e.Dirty
}

func (e *Editor) moveCursorAfterInsert(text string) {
	lines := strings.Split(text, "\n")
	if len(lines) == 1 {
		e.Cursor.Column += len(lines[0])
	} else {
		e.Cursor.Line += len(lines) - 1
		e.Cursor.Column = len(lines[len(lines)-1])
	}
}

func (e *Editor) lineNumWidth() int {
	if !e.ShowLineNumbers {
		return 0
	}
	return e.LineNumWidth
}

func (e *Editor) View() string {
	var sb strings.Builder

	startLine, endLine := e.Viewport.VisibleLineRange()
	if endLine > e.Buffer.LineCount() {
		endLine = e.Buffer.LineCount()
	}

	for i := startLine; i < endLine; i++ {
		e.renderLine(&sb, i)
		if i < endLine-1 {
			sb.WriteString("\n")
		}
	}

	for i := endLine - startLine; i < e.Height; i++ {
		if e.ShowLineNumbers {
			sb.WriteString(fmt.Sprintf("%*s  ", e.LineNumWidth-1, "~"))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (e *Editor) renderLine(sb *strings.Builder, lineNum int) {
	if e.ShowLineNumbers {
		lineNumStr := fmt.Sprintf("%*d ", e.LineNumWidth-1, lineNum+1)
		if e.Cursor.Line == lineNum && e.focused {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(lineNumStr))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(lineNumStr))
		}
	}

	line := ""
	if lineNum < e.Buffer.LineCount() {
		line = e.Buffer.Lines()[lineNum]
	}

	startCol, _ := e.Viewport.VisibleColumnRange()

	lineStart := e.lineOffset(lineNum)
	lineEnd := lineStart + len(line)

	line = e.applyHighlighting(line, lineStart, lineEnd, startCol)

	if len(line) > e.Viewport.Width {
		line = line[:e.Viewport.Width]
	}

	if e.hasSelection() && e.isLineInSelection(lineNum) {
		line = e.renderLineWithSelectionRaw(line, lineNum, startCol)
	}

	if e.Cursor.Line == lineNum && e.ShowCursor && e.focused {
		line = e.renderLineWithCursor(line, lineNum, startCol)
	}

	sb.WriteString(line)
}

func (e *Editor) lineOffset(lineNum int) int {
	offset := 0
	lines := e.Buffer.Lines()
	for i := 0; i < lineNum && i < len(lines); i++ {
		offset += len(lines[i]) + 1
	}
	return offset
}

func (e *Editor) applyHighlighting(line string, lineStart, lineEnd, startCol int) string {
	if len(e.highlightSpans) == 0 {
		if startCol > 0 && startCol < len(line) {
			return line[startCol:]
		} else if startCol >= len(line) {
			return ""
		}
		return line
	}

	var result strings.Builder
	linePos := 0

	for _, span := range e.highlightSpans {
		if span.End <= lineStart {
			continue
		}
		if span.Start >= lineEnd {
			break
		}

		spanStartInLine := span.Start - lineStart
		spanEndInLine := span.End - lineStart

		if spanStartInLine < 0 {
			spanStartInLine = 0
		}
		if spanEndInLine > len(line) {
			spanEndInLine = len(line)
		}

		if spanStartInLine >= len(line) {
			continue
		}

		if spanStartInLine > linePos {
			result.WriteString(line[linePos:spanStartInLine])
		}

		text := line[spanStartInLine:spanEndInLine]
		style := e.theme.StyleForToken(span.TokenType)
		result.WriteString(style.Render(text))

		linePos = spanEndInLine
	}

	if linePos < len(line) {
		result.WriteString(line[linePos:])
	}

	highlighted := result.String()

	if startCol > 0 && startCol < len(highlighted) {
		return highlighted[startCol:]
	} else if startCol >= len(highlighted) {
		return ""
	}
	return highlighted
}

func (e *Editor) isLineInSelection(lineNum int) bool {
	norm := e.Selection.Normalized()
	return lineNum >= norm.Start.Line && lineNum <= norm.End.Line
}

func (e *Editor) renderLineWithSelectionRaw(line string, lineNum, startCol int) string {
	norm := e.Selection.Normalized()

	start := 0
	end := len(line)

	if lineNum == norm.Start.Line {
		start = max(0, norm.Start.Column-startCol)
	}
	if lineNum == norm.End.Line {
		end = min(len(line), norm.End.Column-startCol)
	}

	if start >= end {
		return line
	}

	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color(e.SelectionColor))
	return line[:start] + highlightStyle.Render(line[start:end]) + line[end:]
}

func (e *Editor) renderLineWithCursor(line string, lineNum, startCol int) string {
	cursorCol := e.Cursor.Column - startCol
	if cursorCol < 0 || cursorCol > len(line) {
		return line
	}

	if cursorCol == len(line) {
		return line + e.renderCursor(" ")
	}

	return line[:cursorCol] + e.renderCursor(string(line[cursorCol])) + line[cursorCol+1:]
}

func (e *Editor) renderCursor(char string) string {
	switch e.CursorStyle {
	case CursorBlock:
		return lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).Foreground(lipgloss.Color("#000000")).Render(char)
	case CursorLine:
		return lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).Render(" ") + char
	case CursorUnderline:
		return lipgloss.NewStyle().Underline(true).Render(char)
	}
	return char
}
