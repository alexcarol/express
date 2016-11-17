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

	ast, err := createNumericAST(tokens, 0)
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

type additionNode struct {
	leftNum, rightNum numericNode
}

func (n additionNode) Eval(variables map[string]float64) (float64, error) {
	leftNum, err := n.leftNum.Eval(variables)
	if err != nil {
		return 0, err
	}

	rightNum, err := n.rightNum.Eval(variables)
	if err != nil {
		return 0, err
	}

	return leftNum + rightNum, nil
}

type subtractionNode struct {
	leftNum, rightNum numericNode
}

func (n subtractionNode) Eval(variables map[string]float64) (float64, error) {
	leftNum, err := n.leftNum.Eval(variables)
	if err != nil {
		return 0, err
	}

	rightNum, err := n.rightNum.Eval(variables)
	if err != nil {
		return 0, err
	}

	return leftNum - rightNum, nil
}

func createNumericAST(tokens []token, position int) (numericNode, error) {
	if len(tokens) <= position {
		return nil, eoi{}
	}
	t := tokens[position]
	switch t.kind {
	case lNumber:
		value, err := strconv.ParseFloat(t.text, 64)
		if err != nil {
			return nil, err
		}
		position++

		leftNode := lNumeric{value}

		if len(tokens) <= position {
			return leftNode, nil
		}

		switch tokens[position].kind {
		case plus:
			rightNode, err := createNumericAST(tokens, position+1)

			return additionNode{leftNode, rightNode}, err
		case minus:
			rightNode, err := createNumericAST(tokens, position+1)

			return subtractionNode{leftNode, rightNode}, err
		default:
			return leftNode, unexpectedToken{tokens[position], position, "numeric operator"}
		}

	case lParen:
		node, err := createNumericAST(tokens, position+1)

		utErr, ok := err.(unexpectedToken)
		if ok && utErr.t.kind == rParen {
			return node, nil
		}

		return node, err
	default:
		return nil, unexpectedToken{t, position, ""}
	}
}
