package syntax

import (
	"regexp"
	"sort"
)

type RegexHighlighter struct {
	patterns []pattern
}

type pattern struct {
	regex     *regexp.Regexp
	tokenType TokenType
}

func NewRegexHighlighter(patterns []pattern) *RegexHighlighter {
	return &RegexHighlighter{patterns: patterns}
}

func (h *RegexHighlighter) Highlight(code string) []HighlightSpan {
	var spans []HighlightSpan

	for _, p := range h.patterns {
		matches := p.regex.FindAllStringIndex(code, -1)
		for _, m := range matches {
			spans = append(spans, HighlightSpan{
				Start:     m[0],
				End:       m[1],
				TokenType: p.tokenType,
			})
		}
	}

	sort.Slice(spans, func(i, j int) bool {
		return spans[i].Start < spans[j].Start
	})

	return mergeSpans(spans)
}

func mergeSpans(spans []HighlightSpan) []HighlightSpan {
	if len(spans) == 0 {
		return spans
	}

	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Start != spans[j].Start {
			return spans[i].Start < spans[j].Start
		}
		return spans[i].End > spans[j].End
	})

	var result []HighlightSpan
	for _, span := range spans {
		overlaps := false
		for i := range result {
			if span.Start >= result[i].Start && span.End <= result[i].End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			result = append(result, span)
		}
	}

	return result
}

func NewPythonHighlighter() *RegexHighlighter {
	patterns := []pattern{
		{regexp.MustCompile(`#.*$`), TokenComment},
		{regexp.MustCompile(`""".*?"""|'''.*?'''`), TokenString},
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'`), TokenString},
		{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?\b`), TokenNumber},
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber},
		{regexp.MustCompile(`\b0[oO][0-7]+\b`), TokenNumber},
		{regexp.MustCompile(`\b0[bB][01]+\b`), TokenNumber},
		{regexp.MustCompile(`\b(and|as|assert|async|await|break|class|continue|def|del|elif|else|except|finally|for|from|global|if|import|in|is|lambda|nonlocal|not|or|pass|raise|return|try|while|with|yield)\b`), TokenKeyword},
		{regexp.MustCompile(`\b(True|False|None)\b`), TokenConstant},
		{regexp.MustCompile(`\b(int|float|str|bool|list|dict|set|tuple|bytes|bytearray|memoryview|range|frozenset|type|object|complex)\b`), TokenTypeName},
		{regexp.MustCompile(`\b(print|len|range|input|open|type|isinstance|issubclass|hasattr|getattr|setattr|delattr|callable|super|property|classmethod|staticmethod|enumerate|zip|map|filter|sorted|reversed|any|all|min|max|sum|abs|round|pow|divmod|hex|oct|bin|ord|chr|repr|str|int|float|bool|list|dict|set|tuple|iter|next|slice|format|vars|dir|help|id|hash|exec|eval|compile|globals|locals|breakpoint)\b`), TokenBuiltin},
		{regexp.MustCompile(`\bdef\s+(\w+)`), TokenFunction},
		{regexp.MustCompile(`\bclass\s+(\w+)`), TokenTypeName},
		{regexp.MustCompile(`\b([A-Z][a-zA-Z0-9]*)\b`), TokenTypeName},
		{regexp.MustCompile(`\bself\b`), TokenVariable},
		{regexp.MustCompile(`[\+\-\*/%=<>!&|^~]+`), TokenOperator},
		{regexp.MustCompile(`[\(\)\[\]\{\},;:\.]`), TokenPunctuation},
	}

	return NewRegexHighlighter(patterns)
}

func NewGoHighlighter() *RegexHighlighter {
	patterns := []pattern{
		{regexp.MustCompile(`//.*$`), TokenComment},
		{regexp.MustCompile(`/\*[\s\S]*?\*/`), TokenComment},
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"`), TokenString},
		{regexp.MustCompile("`[^`]*`"), TokenString},
		{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?\b`), TokenNumber},
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber},
		{regexp.MustCompile(`\b0[oO][0-7]+\b`), TokenNumber},
		{regexp.MustCompile(`\b(break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b`), TokenKeyword},
		{regexp.MustCompile(`\b(true|false|nil|iota)\b`), TokenConstant},
		{regexp.MustCompile(`\b(bool|byte|complex64|complex128|error|float32|float64|int|int8|int16|int32|int64|rune|string|uint|uint8|uint16|uint32|uint64|uintptr)\b`), TokenTypeName},
		{regexp.MustCompile(`\bfunc\s+(\w+)`), TokenFunction},
		{regexp.MustCompile(`\btype\s+(\w+)\s+struct`), TokenTypeName},
		{regexp.MustCompile(`\b[A-Z][a-zA-Z0-9]*\b`), TokenTypeName},
		{regexp.MustCompile(`[\+\-\*/%=<>!&|^~:]+`), TokenOperator},
		{regexp.MustCompile(`[\(\)\[\]\{\},;]`), TokenPunctuation},
	}

	return NewRegexHighlighter(patterns)
}

func NewJSHighlighter() *RegexHighlighter {
	patterns := []pattern{
		{regexp.MustCompile(`//.*$`), TokenComment},
		{regexp.MustCompile(`/\*[\s\S]*?\*/`), TokenComment},
		{regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'`), TokenString},
		{regexp.MustCompile("`[^`]*`"), TokenString},
		{regexp.MustCompile(`\b[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?\b`), TokenNumber},
		{regexp.MustCompile(`\b0[xX][0-9a-fA-F]+\b`), TokenNumber},
		{regexp.MustCompile(`\b(break|case|catch|const|continue|debugger|default|delete|do|else|export|extends|finally|for|function|if|import|in|instanceof|let|new|return|super|switch|this|throw|try|typeof|var|void|while|with|yield|class|enum|await|async|static|get|set)\b`), TokenKeyword},
		{regexp.MustCompile(`\b(true|false|null|undefined|NaN|Infinity)\b`), TokenConstant},
		{regexp.MustCompile(`\b(Array|Boolean|Date|Function|Map|Number|Object|Promise|RegExp|Set|String|Symbol|WeakMap|WeakSet|Error|console|document|window)\b`), TokenBuiltin},
		{regexp.MustCompile(`\bfunction\s+(\w+)`), TokenFunction},
		{regexp.MustCompile(`\bclass\s+(\w+)`), TokenTypeName},
		{regexp.MustCompile(`[\+\-\*/%=<>!&|^~?]+`), TokenOperator},
		{regexp.MustCompile(`[\(\)\[\]\{\},;:.]`), TokenPunctuation},
	}

	return NewRegexHighlighter(patterns)
}

func init() {
	RegisterLanguage(".go", NewGoHighlighter())
	RegisterLanguage(".js", NewJSHighlighter())
	RegisterLanguage(".mjs", NewJSHighlighter())
	RegisterLanguage(".cjs", NewJSHighlighter())
}
