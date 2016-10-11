package express

import (
	"errors"
	"strconv"
)

// NumericEval returns the result for an expression with provided variables
func NumericEval(expression string, variables map[string]float64) (float64, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return 0, err
	}

	ast, err := createNumericAST(tokens)
	if err != nil {
		if _, ok := err.(eoi); !ok {
			return 0, err
		}
	}

	if ast == nil {
		return 0, errors.New("unexpected nil ast")
	}

	return ast.Eval(variables)
}

type numericNode interface {
	Eval(map[string]float64) (float64, error)
}

type lNumeric struct {
	value float64
}

func (n lNumeric) Eval(map[string]float64) (float64, error) {
	return n.value, nil
}

func createNumericAST(tokens []token) (numericNode, error) {
	node, _, err := numericNodeStartingAt(tokens, 0)

	return node, err
}

func numericNodeStartingAt(tokens []token, position int) (numericNode, int, error) {
	if len(tokens) <= position {
		return nil, position, eoi{}
	}
	t := tokens[position]
	switch t.kind {
	case lNumber:
		value, err := strconv.ParseFloat(t.text, 64)

		return lNumeric{value}, position + 1, err
	case lParen:
		node, err := createNumericAST(tokens[position+1:])

		utErr, ok := err.(unexpectedToken)
		if ok && utErr.t.kind == rParen {
			return node, utErr.position + 1, nil
		}

		return node, position, err
	default:
		return nil, position, unexpectedToken{t, position}
	}
}
