package buffer

type Action interface {
	Apply(b *Buffer)
	Undo(b *Buffer)
}

type InsertAction struct {
	Line     int
	Col      int
	Char     rune
	IsGroup  bool
}

func (a *InsertAction) Apply(b *Buffer) {
	b.insertAt(a.Line, a.Col, a.Char)
}

func (a *InsertAction) Undo(b *Buffer) {
	b.deleteAt(a.Line, a.Col)
}

type InsertStringAction struct {
	Line    int
	Col     int
	Content string
}

func (a *InsertStringAction) Apply(b *Buffer) {
	b.insertStringAt(a.Line, a.Col, a.Content)
}

func (a *InsertStringAction) Undo(b *Buffer) {
	line, col := a.Line, a.Col
	for range a.Content {
		line, col = b.deleteAt(line, col)
	}
}

type DeleteAction struct {
	Line     int
	Col      int
	Char     rune
	IsGroup  bool
}

func (a *DeleteAction) Apply(b *Buffer) {
	b.deleteAt(a.Line, a.Col)
}

func (a *DeleteAction) Undo(b *Buffer) {
	b.insertAt(a.Line, a.Col, a.Char)
}

type DeleteLineAction struct {
	Line    int
	Content string
}

func (a *DeleteLineAction) Apply(b *Buffer) {
	b.deleteLineAt(a.Line)
}

func (a *DeleteLineAction) Undo(b *Buffer) {
	b.insertLineAt(a.Line, a.Content)
}

type BackspaceAction struct {
	Line     int
	Col      int
	Char     rune
	IsGroup  bool
}

func (a *BackspaceAction) Apply(b *Buffer) {
	b.backspaceAt(a.Line, a.Col)
}

func (a *BackspaceAction) Undo(b *Buffer) {
	b.insertAt(a.Line, a.Col-1, a.Char)
}

type ActionGroup struct {
	Actions []Action
}

func (g *ActionGroup) Apply(b *Buffer) {
	for _, a := range g.Actions {
		a.Apply(b)
	}
}

func (g *ActionGroup) Undo(b *Buffer) {
	for i := len(g.Actions) - 1; i >= 0; i-- {
		g.Actions[i].Undo(b)
	}
}

func canGroup(a1, a2 Action) bool {
	switch v1 := a1.(type) {
	case *InsertAction:
		if v2, ok := a2.(*InsertAction); ok {
			return v1.IsGroup && v2.IsGroup
		}
	case *DeleteAction:
		if v2, ok := a2.(*DeleteAction); ok {
			return v1.IsGroup && v2.IsGroup
		}
	case *BackspaceAction:
		if v2, ok := a2.(*BackspaceAction); ok {
			return v1.IsGroup && v2.IsGroup
		}
	}
	return false
}
