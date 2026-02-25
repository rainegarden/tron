package buffer

import (
	"os"
	"strings"
	"sync"
)

type Position struct {
	Line int
	Col  int
}

type Selection struct {
	Start Position
	End   Position
}

type Buffer struct {
	mu       sync.RWMutex
	lines    []string
	cursor   Position
	selection *Selection
	undoStack []Action
	redoStack []Action
	filePath string
	dirty    bool
	grouping bool
	group    *ActionGroup
}

func NewBuffer() *Buffer {
	return &Buffer{
		lines: []string{""},
	}
}

func NewBufferFromFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	b := &Buffer{
		lines:    lines,
		filePath: path,
	}
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	return b, nil
}

func (b *Buffer) Save() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.filePath == "" {
		return os.ErrInvalid
	}

	content := strings.Join(b.lines, "\n")
	err := os.WriteFile(b.filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	b.dirty = false
	return nil
}

func (b *Buffer) SaveAs(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	content := strings.Join(b.lines, "\n")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return err
	}

	b.filePath = path
	b.dirty = false
	return nil
}

func (b *Buffer) Insert(char rune) {
	b.mu.Lock()
	defer b.mu.Unlock()

	line := b.cursor.Line
	col := b.cursor.Col

	b.insertAt(line, col, char)
	b.cursor.Col++

	action := &InsertAction{Line: line, Col: col, Char: char, IsGroup: true}
	b.pushAction(action)
	b.dirty = true
}

func (b *Buffer) InsertString(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	line := b.cursor.Line
	col := b.cursor.Col

	b.insertStringAt(line, col, s)

	lines := strings.Split(s, "\n")
	lineCount := len(lines) - 1
	if lineCount > 0 {
		b.cursor.Line += lineCount
		b.cursor.Col = len(lines[lineCount])
	} else {
		b.cursor.Col += len(s)
	}

	action := &InsertStringAction{Line: line, Col: col, Content: s}
	b.pushAction(action)
	b.dirty = true
}

func (b *Buffer) Delete() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Line >= len(b.lines) {
		return
	}

	line := b.cursor.Line
	col := b.cursor.Col
	currentLine := b.lines[line]

	if col < len(currentLine) {
		char := rune(currentLine[col])
		b.deleteAt(line, col)
		action := &DeleteAction{Line: line, Col: col, Char: char, IsGroup: true}
		b.pushAction(action)
		b.dirty = true
	} else if line < len(b.lines)-1 {
		b.lines[line] = currentLine + b.lines[line+1]
		b.lines = append(b.lines[:line+1], b.lines[line+2:]...)
		action := &DeleteAction{Line: line, Col: col, Char: '\n', IsGroup: false}
		b.pushAction(action)
		b.dirty = true
	}
}

func (b *Buffer) Backspace() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Line == 0 && b.cursor.Col == 0 {
		return
	}

	line := b.cursor.Line
	col := b.cursor.Col

	if col > 0 {
		b.cursor.Col--
		char := rune(b.lines[line][col-1])
		b.backspaceAt(line, col)
		action := &BackspaceAction{Line: line, Col: col, Char: char, IsGroup: true}
		b.pushAction(action)
		b.dirty = true
	} else if line > 0 {
		prevLineLen := len(b.lines[line-1])
		b.lines[line-1] += b.lines[line]
		b.lines = append(b.lines[:line], b.lines[line+1:]...)
		b.cursor.Line--
		b.cursor.Col = prevLineLen
		action := &BackspaceAction{Line: line, Col: col, Char: '\n', IsGroup: false}
		b.pushAction(action)
		b.dirty = true
	}
}

func (b *Buffer) DeleteLine() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Line >= len(b.lines) {
		return
	}

	content := b.lines[b.cursor.Line]
	b.deleteLineAt(b.cursor.Line)

	if len(b.lines) == 0 {
		b.lines = []string{""}
	}

	if b.cursor.Line >= len(b.lines) {
		b.cursor.Line = len(b.lines) - 1
	}
	b.cursor.Col = 0

	action := &DeleteLineAction{Line: b.cursor.Line, Content: content}
	b.pushAction(action)
	b.dirty = true
}

func (b *Buffer) MoveUp() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Line > 0 {
		b.cursor.Line--
		b.clampCursor()
	}
}

func (b *Buffer) MoveDown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Line < len(b.lines)-1 {
		b.cursor.Line++
		b.clampCursor()
	}
}

func (b *Buffer) MoveLeft() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Col > 0 {
		b.cursor.Col--
	} else if b.cursor.Line > 0 {
		b.cursor.Line--
		b.cursor.Col = len(b.lines[b.cursor.Line])
	}
}

func (b *Buffer) MoveRight() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cursor.Col < len(b.lines[b.cursor.Line]) {
		b.cursor.Col++
	} else if b.cursor.Line < len(b.lines)-1 {
		b.cursor.Line++
		b.cursor.Col = 0
	}
}

func (b *Buffer) MoveToLineStart() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cursor.Col = 0
}

func (b *Buffer) MoveToLineEnd() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cursor.Col = len(b.lines[b.cursor.Line])
}

func (b *Buffer) MoveToStart() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cursor = Position{Line: 0, Col: 0}
}

func (b *Buffer) MoveToEnd() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cursor.Line = len(b.lines) - 1
	b.cursor.Col = len(b.lines[b.cursor.Line])
}

func (b *Buffer) Undo() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.undoStack) == 0 {
		return
	}

	action := b.undoStack[len(b.undoStack)-1]
	b.undoStack = b.undoStack[:len(b.undoStack)-1]
	action.Undo(b)
	b.redoStack = append(b.redoStack, action)
	b.dirty = true
}

func (b *Buffer) Redo() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.redoStack) == 0 {
		return
	}

	action := b.redoStack[len(b.redoStack)-1]
	b.redoStack = b.redoStack[:len(b.redoStack)-1]
	action.Apply(b)
	b.undoStack = append(b.undoStack, action)
	b.dirty = true
}

func (b *Buffer) LineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines)
}

func (b *Buffer) GetLine(n int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if n < 0 || n >= len(b.lines) {
		return ""
	}
	return b.lines[n]
}

func (b *Buffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return strings.Join(b.lines, "\n")
}

func (b *Buffer) Cursor() Position {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cursor
}

func (b *Buffer) SetCursor(line, col int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cursor.Line = line
	b.cursor.Col = col
	b.clampCursor()
}

func (b *Buffer) FilePath() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.filePath
}

func (b *Buffer) IsDirty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dirty
}

func (b *Buffer) Selection() *Selection {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.selection
}

func (b *Buffer) insertAt(line, col int, char rune) {
	if line < 0 || line >= len(b.lines) {
		return
	}
	currentLine := b.lines[line]
	if col < 0 || col > len(currentLine) {
		return
	}

	str := string(char)
	b.lines[line] = currentLine[:col] + str + currentLine[col:]
}

func (b *Buffer) insertStringAt(line, col int, s string) {
	if line < 0 || line >= len(b.lines) {
		return
	}
	currentLine := b.lines[line]
	if col < 0 || col > len(currentLine) {
		return
	}

	parts := strings.Split(s, "\n")
	if len(parts) == 1 {
		b.lines[line] = currentLine[:col] + s + currentLine[col:]
		return
	}

	b.lines[line] = currentLine[:col] + parts[0]
	newLines := make([]string, 0, len(b.lines)+len(parts)-1)
	newLines = append(newLines, b.lines[:line+1]...)
	for i := 1; i < len(parts)-1; i++ {
		newLines = append(newLines, parts[i])
	}
	newLines = append(newLines, parts[len(parts)-1]+currentLine[col:])
	newLines = append(newLines, b.lines[line+1:]...)
	b.lines = newLines
}

func (b *Buffer) deleteAt(line, col int) Position {
	if line < 0 || line >= len(b.lines) {
		return b.cursor
	}
	currentLine := b.lines[line]
	if col < 0 || col >= len(currentLine) {
		return b.cursor
	}

	b.lines[line] = currentLine[:col] + currentLine[col+1:]
	return Position{Line: line, Col: col}
}

func (b *Buffer) backspaceAt(line, col int) {
	if line < 0 || line >= len(b.lines) || col <= 0 {
		return
	}
	currentLine := b.lines[line]
	b.lines[line] = currentLine[:col-1] + currentLine[col:]
}

func (b *Buffer) deleteLineAt(line int) {
	if line < 0 || line >= len(b.lines) {
		return
	}
	b.lines = append(b.lines[:line], b.lines[line+1:]...)
}

func (b *Buffer) insertLineAt(line int, content string) {
	if line < 0 || line > len(b.lines) {
		return
	}
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:line]...)
	newLines = append(newLines, content)
	newLines = append(newLines, b.lines[line:]...)
	b.lines = newLines
}

func (b *Buffer) clampCursor() {
	if b.cursor.Line < 0 {
		b.cursor.Line = 0
	}
	if b.cursor.Line >= len(b.lines) {
		b.cursor.Line = len(b.lines) - 1
	}
	if b.cursor.Col < 0 {
		b.cursor.Col = 0
	}
	lineLen := len(b.lines[b.cursor.Line])
	if b.cursor.Col > lineLen {
		b.cursor.Col = lineLen
	}
}

func (b *Buffer) pushAction(action Action) {
	if b.grouping && b.group != nil {
		b.group.Actions = append(b.group.Actions, action)
		return
	}

	if len(b.undoStack) > 0 && canGroup(b.undoStack[len(b.undoStack)-1], action) {
		if g, ok := b.undoStack[len(b.undoStack)-1].(*ActionGroup); ok {
			g.Actions = append(g.Actions, action)
			return
		}
		prev := b.undoStack[len(b.undoStack)-1]
		g := &ActionGroup{Actions: []Action{prev, action}}
		b.undoStack[len(b.undoStack)-1] = g
		return
	}

	b.undoStack = append(b.undoStack, action)
	b.redoStack = b.redoStack[:0]
}

func (b *Buffer) BeginGroup() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.grouping = true
	b.group = &ActionGroup{}
}

func (b *Buffer) EndGroup() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.grouping && b.group != nil && len(b.group.Actions) > 0 {
		if len(b.group.Actions) == 1 {
			b.undoStack = append(b.undoStack, b.group.Actions[0])
		} else {
			b.undoStack = append(b.undoStack, b.group)
		}
		b.redoStack = b.redoStack[:0]
	}
	b.grouping = false
	b.group = nil
}

func (b *Buffer) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.undoStack = nil
	b.redoStack = nil
}
