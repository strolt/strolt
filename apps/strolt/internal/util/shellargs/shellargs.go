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

// Split tokenizes s. Callers that validated their input earlier use
// MustSplit instead.
func Split(s string) ([]string, error) {
	runes := []rune(s)

	var (
		args     []string
		current  []rune
		hasToken bool
		state    quoteState
	)

	flush := func() {
		if hasToken {
			args = append(args, string(current))
			current = nil
			hasToken = false
		}
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch state {
		case quoteSingle:
			if r == '\'' {
				state = quoteNone
				continue
			}

			current = append(current, r)
		case quoteDouble:
			switch {
			case r == '"':
				state = quoteNone
			case r == '\\' && i+1 < len(runes) && (runes[i+1] == '"' || runes[i+1] == '\\'):
				i++
				current = append(current, runes[i])
			default:
				current = append(current, r)
			}
		default:
			switch {
			case r == ' ' || r == '\t' || r == '\n' || r == '\r':
				flush()
			case r == '\'':
				hasToken = true
				state = quoteSingle
			case r == '"':
				hasToken = true
				state = quoteDouble
			case r == '\\' && i+1 < len(runes):
				hasToken = true
				i++
				current = append(current, runes[i])
			default:
				hasToken = true
				current = append(current, r)
			}
		}
	}

	if state != quoteNone {
		return nil, fmt.Errorf("unterminated %s quote", quoteName(state))
	}

	flush()

	return args, nil
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
