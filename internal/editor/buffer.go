package editor

type CursorStyle int

const (
	CursorBlock CursorStyle = iota
	CursorLine
	CursorUnderline
)

type Position struct {
	Line   int
	Column int
}

type Selection struct {
	Start Position
	End   Position
}

func (s Selection) IsEmpty() bool {
	return s.Start.Line == s.End.Line && s.Start.Column == s.End.Column
}

func (s Selection) Normalized() Selection {
	if s.Start.Line > s.End.Line || (s.Start.Line == s.End.Line && s.Start.Column > s.End.Column) {
		return Selection{Start: s.End, End: s.Start}
	}
	return s
}

type Buffer interface {
	Content() string
	Lines() []string
	LineCount() int
	LineLength(line int) int
	CharAt(line, col int) rune
	Insert(pos Position, text string)
	Delete(start, end Position)
	DeleteChar(pos Position, forward bool)
	GetText(start, end Position) string
	SetContent(content string)
}

type SimpleBuffer struct {
	lines []string
}

func NewSimpleBuffer() *SimpleBuffer {
	return &SimpleBuffer{lines: []string{""}}
}

func NewSimpleBufferWithContent(content string) *SimpleBuffer {
	b := &SimpleBuffer{}
	b.SetContent(content)
	return b
}

func (b *SimpleBuffer) Content() string {
	result := ""
	for i, line := range b.lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func (b *SimpleBuffer) Lines() []string {
	return b.lines
}

func (b *SimpleBuffer) LineCount() int {
	return len(b.lines)
}

func (b *SimpleBuffer) LineLength(line int) int {
	if line < 0 || line >= len(b.lines) {
		return 0
	}
	return len(b.lines[line])
}

func (b *SimpleBuffer) CharAt(line, col int) rune {
	if line < 0 || line >= len(b.lines) {
		return 0
	}
	if col < 0 || col >= len(b.lines[line]) {
		return 0
	}
	return rune(b.lines[line][col])
}

func (b *SimpleBuffer) Insert(pos Position, text string) {
	for pos.Line >= len(b.lines) {
		b.lines = append(b.lines, "")
	}

	if text == "\n" || text == "\r\n" {
		currentLine := b.lines[pos.Line]
		before := currentLine[:min(pos.Column, len(currentLine))]
		after := ""
		if pos.Column < len(currentLine) {
			after = currentLine[pos.Column:]
		}
		b.lines = append(b.lines, "")
		copy(b.lines[pos.Line+2:], b.lines[pos.Line+1:])
		b.lines[pos.Line] = before
		b.lines[pos.Line+1] = after
		return
	}

	currentLine := b.lines[pos.Line]
	col := min(pos.Column, len(currentLine))
	b.lines[pos.Line] = currentLine[:col] + text + currentLine[col:]
}

func (b *SimpleBuffer) Delete(start, end Position) {
	start, end = normalizeRange(start, end)

	if start.Line == end.Line {
		if start.Line < len(b.lines) {
			line := b.lines[start.Line]
			b.lines[start.Line] = line[:start.Column] + line[end.Column:]
		}
		return
	}

	if start.Line >= len(b.lines) || end.Line >= len(b.lines) {
		return
	}

	firstLine := b.lines[start.Line]
	lastLine := b.lines[end.Line]
	newLine := firstLine[:start.Column] + lastLine[end.Column:]

	newLines := make([]string, 0, len(b.lines)-(end.Line-start.Line))
	newLines = append(newLines, b.lines[:start.Line]...)
	newLines = append(newLines, newLine)
	newLines = append(newLines, b.lines[end.Line+1:]...)

	b.lines = newLines
}

func (b *SimpleBuffer) DeleteChar(pos Position, forward bool) {
	if pos.Line < 0 || pos.Line >= len(b.lines) {
		return
	}

	line := b.lines[pos.Line]

	if forward {
		if pos.Column < len(line) {
			b.lines[pos.Line] = line[:pos.Column] + line[pos.Column+1:]
		} else if pos.Line < len(b.lines)-1 {
			b.lines[pos.Line] = line + b.lines[pos.Line+1]
			b.lines = append(b.lines[:pos.Line+1], b.lines[pos.Line+2:]...)
		}
	} else {
		if pos.Column > 0 {
			b.lines[pos.Line] = line[:pos.Column-1] + line[pos.Column:]
		} else if pos.Line > 0 {
			prevLine := b.lines[pos.Line-1]
			b.lines[pos.Line-1] = prevLine + line
			b.lines = append(b.lines[:pos.Line], b.lines[pos.Line+1:]...)
		}
	}
}

func (b *SimpleBuffer) GetText(start, end Position) string {
	start, end = normalizeRange(start, end)

	if start.Line == end.Line {
		if start.Line < len(b.lines) {
			line := b.lines[start.Line]
			if start.Column < len(line) && end.Column <= len(line) {
				return line[start.Column:end.Column]
			}
		}
		return ""
	}

	var result string
	if start.Line < len(b.lines) {
		result = b.lines[start.Line][min(start.Column, len(b.lines[start.Line])):] + "\n"
	}

	for i := start.Line + 1; i < end.Line && i < len(b.lines); i++ {
		result += b.lines[i] + "\n"
	}

	if end.Line < len(b.lines) {
		result += b.lines[end.Line][:min(end.Column, len(b.lines[end.Line]))]
	}

	return result
}

func (b *SimpleBuffer) SetContent(content string) {
	if content == "" {
		b.lines = []string{""}
		return
	}

	lines := []string{}
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	lines = append(lines, content[start:])
	b.lines = lines
}

func normalizeRange(start, end Position) (Position, Position) {
	if start.Line > end.Line || (start.Line == end.Line && start.Column > end.Column) {
		return end, start
	}
	return start, end
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
