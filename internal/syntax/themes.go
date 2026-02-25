package syntax

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Keyword    lipgloss.Style
	String     lipgloss.Style
	Comment    lipgloss.Style
	Number     lipgloss.Style
	Function   lipgloss.Style
	Operator   lipgloss.Style
	Identifier lipgloss.Style
	Type       lipgloss.Style
	Builtin    lipgloss.Style
	Constant   lipgloss.Style
	Variable   lipgloss.Style
	Punctuation lipgloss.Style
}

func DefaultTheme() *Theme {
	return &Theme{
		Keyword: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")),
		String:  lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")),
		Comment: lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")),
		Number:  lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")),
		Function: lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")),
		Operator: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")),
		Identifier: lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")),
		Type:    lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")),
		Builtin: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")),
		Constant: lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")),
		Variable: lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")),
		Punctuation: lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2")),
	}
}

func (t *Theme) StyleForToken(tt TokenType) lipgloss.Style {
	switch tt {
	case TokenKeyword:
		return t.Keyword
	case TokenString:
		return t.String
	case TokenComment:
		return t.Comment
	case TokenNumber:
		return t.Number
	case TokenFunction:
		return t.Function
	case TokenOperator:
		return t.Operator
	case TokenIdentifier:
		return t.Identifier
	case TokenTypeName:
		return t.Type
	case TokenBuiltin:
		return t.Builtin
	case TokenConstant:
		return t.Constant
	case TokenVariable:
		return t.Variable
	case TokenPunctuation:
		return t.Punctuation
	default:
		return lipgloss.NewStyle()
	}
}

var defaultTheme = DefaultTheme()

func GetTheme() *Theme {
	return defaultTheme
}
