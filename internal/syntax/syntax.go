// Package syntax provides syntax highlighting for the editor.
//
// Current implementation uses regex-based highlighting (see regex.go) which is
// portable and doesn't require CGO. This file defines the core interfaces.
//
// To add tree-sitter support for more accurate parsing:
//
//  1. Install tree-sitter: go get github.com/tree-sitter/go-tree-sitter
//  2. For each language, you need grammar files. Options:
//     a) Use pre-built shared libraries (requires CGO_ENABLED=1)
//     b) Build grammars at compile-time using go:generate
//
//  Example tree-sitter implementation:
//
//     import sitter "github.com/tree-sitter/go-tree-sitter"
//
//     type TreeSitterHighlighter struct {
//         parser *sitter.Parser
//         language *sitter.Language
//         queries map[string]*sitter.Query  // highlight queries per language
//     }
//
//     func NewTreeSitterHighlighter(lang *sitter.Language, query string) *TreeSitterHighlighter {
//         parser := sitter.NewParser()
//         parser.SetLanguage(lang)
//         q, _ := sitter.NewQuery([]byte(query), lang)
//         return &TreeSitterHighlighter{parser: parser, language: lang, queries: query}
//     }
//
//  To add a new language with regex highlighting, add to regex.go:
//
//     func NewRustHighlighter() *RegexHighlighter {
//         patterns := []pattern{
//             {regexp.MustCompile(`//.*$`), TokenComment},
//             // ... more patterns
//         }
//         return NewRegexHighlighter(patterns)
//     }
//
//     func init() {
//         RegisterLanguage(".rs", NewRustHighlighter())
//     }
package syntax

type TokenType int

const (
	TokenNone TokenType = iota
	TokenKeyword
	TokenString
	TokenComment
	TokenNumber
	TokenFunction
	TokenOperator
	TokenIdentifier
	TokenTypeName
	TokenBuiltin
	TokenConstant
	TokenVariable
	TokenPunctuation
)

type HighlightSpan struct {
	Start     int
	End       int
	TokenType TokenType
}

type Highlighter interface {
	Highlight(code string) []HighlightSpan
}

var languages = make(map[string]Highlighter)

func RegisterLanguage(ext string, h Highlighter) {
	languages[ext] = h
}

func GetHighlighter(ext string) Highlighter {
	if h, ok := languages[ext]; ok {
		return h
	}
	return nil
}

func Highlight(code string, ext string) []HighlightSpan {
	h := GetHighlighter(ext)
	if h == nil {
		return nil
	}
	return h.Highlight(code)
}

func init() {
	RegisterLanguage(".py", NewPythonHighlighter())
	RegisterLanguage(".pyw", NewPythonHighlighter())
}
