package express

import (
	"fmt"
	"strings"
)

const (
	lTrue uint = iota
	lFalse
	not
	and
	or
	variable
	lParen
	rParen
)

type token struct {
	kind uint
	text string
}

func tokenize(expression string) ([]token, error) {
	var tokens []token

	var i = 0

MainLoop:
	for i < len(expression) {
		if expression[i] == byte(' ') { // TODO add other blank spaces
			i++
			continue
		}

		var definedSymbols = map[string]uint{
			"(": lParen,
			")": rParen,
		}

		for text, kind := range definedSymbols {
			if strings.HasPrefix(expression[i:], text) {
				tokens = append(tokens, token{kind, text})
				i += len(text)
				continue MainLoop
			}
		}

		startingI := i
		if isValidStarterIdent(expression[i]) {
			i++
			for i < len(expression) && canBeIdent(expression[i]) {
				i++
			}
		} else {
			return nil, fmt.Errorf("error parsing expression at position %d, found %c", i, expression[i])
		}

		var definedTokens = map[string]uint{
			"not":   not,
			"true":  lTrue,
			"false": lFalse,
			"and":   and,
			"or":    or,
		}

		text := expression[startingI:i]
		kind, found := definedTokens[text]
		if !found {
			kind = variable
		}
		tokens = append(tokens, token{kind, text})
		i++
	}

	return tokens, nil
}

func isValidStarterIdent(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'z'
}

func canBeIdent(b byte) bool {
	return isValidStarterIdent(b)
}
