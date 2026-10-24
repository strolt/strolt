// Package shellargs splits a configuration string into process arguments
// with shell-like quoting semantics. It backs the *_params driver options,
// so a config value like `--opt-a --opt-b --comment="two words"` becomes
// three argv elements instead of one invalid blob.
//
// Rules:
//   - Unquoted whitespace separates arguments.
//   - Single quotes group a value; inside them every character is literal,
//     including backslashes.
//   - Double quotes group a value; inside them a backslash escapes only the
//     double quote and the backslash itself ("\\"" and "\\"), any other
//     backslash stays literal. Variables are not expanded, so "$" is not
//     special.
//   - Outside quotes a backslash escapes the next character (a space stays
//     in the argument, a quote does not open quoting); a trailing backslash
//     with nothing after it stays literal.
//   - An explicitly quoted empty value (” or "") is a valid empty argument
//     and is preserved.
//
// Split reports an error only for an unterminated quote.
package shellargs

import "fmt"

type quoteState int

const (
	quoteNone quoteState = iota
	quoteSingle
	quoteDouble
)

type tokenizer struct {
	args     []string
	current  []rune
	hasToken bool
	state    quoteState
}

// Split tokenizes s. Callers that validated their input earlier use
// MustSplit instead.
func Split(s string) ([]string, error) {
	runes := []rune(s)
	t := &tokenizer{}

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch t.state {
		case quoteSingle:
			t.consumeSingleQuoted(r)
		case quoteDouble:
			i = t.consumeDoubleQuoted(runes, i)
		case quoteNone:
			if r == '\\' && i+1 < len(runes) {
				t.consumeEscaped(runes[i+1])
				i++
				continue
			}

			t.consumeUnquoted(r)
		}
	}

	if t.state != quoteNone {
		return nil, fmt.Errorf("unterminated %s quote", quoteName(t.state))
	}

	t.flush()

	return t.args, nil
}

// consumeSingleQuoted appends r literally; only the closing quote changes
// the state.
func (t *tokenizer) consumeSingleQuoted(r rune) {
	if r == '\'' {
		t.state = quoteNone
		return
	}

	t.current = append(t.current, r)
}

// consumeDoubleQuoted handles one rune inside double quotes and returns the
// (possibly advanced) rune index: a backslash may consume the escaped rune
// next to it.
func (t *tokenizer) consumeDoubleQuoted(runes []rune, i int) int {
	r := runes[i]

	switch {
	case r == '"':
		t.state = quoteNone
	case r == '\\' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\'):
		t.current = append(t.current, runes[i+1])
		return i + 1
	default:
		t.current = append(t.current, r)
	}

	return i
}

// consumeUnquoted handles one rune outside quotes: whitespace ends the
// current argument, quotes open a group, a backslash escapes the next rune.
func (t *tokenizer) consumeUnquoted(r rune) {
	switch {
	case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		t.flush()
	case r == '\'':
		t.hasToken = true
		t.state = quoteSingle
	case r == '"':
		t.hasToken = true
		t.state = quoteDouble
	default:
		t.hasToken = true
		t.current = append(t.current, r)
	}
}

// consumeEscaped appends the rune after a backslash outside quotes.
func (t *tokenizer) consumeEscaped(r rune) {
	t.hasToken = true
	t.current = append(t.current, r)
}

// flush commits the current argument; an explicitly quoted empty value has
// hasToken set without any content and must survive as an empty argument.
func (t *tokenizer) flush() {
	if t.hasToken {
		t.args = append(t.args, string(t.current))
		t.current = nil
		t.hasToken = false
	}
}

// MustSplit is Split for input already validated at configuration load time.
// An error here means validation was bypassed, so it fails loudly instead of
// silently producing a broken argv.
func MustSplit(s string) []string {
	args, err := Split(s)
	if err != nil {
		panic(fmt.Sprintf("shellargs: %s: %q", err, s))
	}

	return args
}

func quoteName(state quoteState) string {
	if state == quoteDouble {
		return "double"
	}

	return "single"
}
