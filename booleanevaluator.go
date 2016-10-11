package express

import (
	"errors"
	"fmt"
)

type unexpectedToken struct {
	t        token
	position int
}

func (e unexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token (%s) of kind %d at position %d", e.t.text, e.t.kind, e.position)
}

type eoi struct{}

func (e eoi) Error() string {
	return "end of input"
}

// BoolEval returns the result for an expression with provided variables
func BoolEval(expression string, variables map[string]interface{}) (b bool, err error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return false, err
	}

	ast, err := createBooleanAST(tokens)
	if err != nil {
		if _, ok := err.(eoi); !ok { // TODO this error probably shouldn't get so far
			return false, err
		}
	}

	if ast == nil {
		return false, errors.New("unexpected nil ast")
	}

	return ast.Eval(variables)
}

type boolNode interface {
	Eval(map[string]interface{}) (bool, error)
}

type lBool struct {
	value bool
}

func (l lBool) Eval(map[string]interface{}) (bool, error) {
	return l.value, nil
}

type varNode struct {
	varName string
}

func (n varNode) Eval(params map[string]interface{}) (bool, error) {
	// TODO we need to check for the variable existing
	v, ok := params[n.varName].(bool)
	if !ok {
		// TODO split in two variables
		return false, fmt.Errorf("variable %s not found or not boolean", n.varName)
	}

	return v, nil
}

type logicalNode struct {
	leftBool, rightBool boolNode
	op                  uint
}

func (l logicalNode) Eval(params map[string]interface{}) (bool, error) {
	// TODO avoid this switch on expression runtime, consider creating interface for EvalLogical()
	switch l.op {
	case and:
		left, err := l.leftBool.Eval(params)
		if err != nil || !left {
			return false, err
		}

		return l.rightBool.Eval(params)
	case or:
		left, err := l.leftBool.Eval(params)
		if left {
			return true, nil
		}

		// TODO consider ignoring errors but logging them
		// or even making it configurable
		// optionally we could check that all
		// the needed variables are set beforehand
		if err != nil {
			return false, err
		}

		return l.rightBool.Eval(params)
	default:
		// The best is probably to make an "andNode", orNode, ...
		return false, fmt.Errorf("unrecognised operator %d", l.op)
	}
}

type notNode struct {
	node boolNode
}

func (n notNode) Eval(parameters map[string]interface{}) (bool, error) {
	res, err := n.node.Eval(parameters)
	if err != nil {
		return false, err
	}

	return !res, nil
}

func boolNodeStartingAt(tokens []token, position int) (boolNode, int, error) {
	if len(tokens) <= position {
		return nil, position, eoi{}
	}
	t := tokens[position]
	switch t.kind {
	case lTrue, lFalse:
		return lBool{t.kind == lTrue}, position + 1, nil
	case variable:
		return varNode{t.text}, position + 1, nil
	case not:
		node, returnPosition, err := boolNodeStartingAt(tokens, position+1)

		return notNode{node}, returnPosition, err
	case lParen:
		node, err := createBooleanAST(tokens[position+1:])

		utErr, ok := err.(unexpectedToken)
		if ok && utErr.t.kind == rParen {
			return node, utErr.position + 1, nil
		}

		return node, position, err
	default:
		return nil, position, unexpectedToken{t, position}
	}
}

func createBooleaASTWithLeftSideAndOperator(left boolNode, operator uint, tokens []token) (boolNode, error) {
	right, i, err := boolNodeStartingAt(tokens, 0)
	if err != nil {
		return logicalNode{left, right, operator}, err
	}

	if i >= len(tokens) {
		return logicalNode{left, right, operator}, eoi{}
	}

	if !isOperator(tokens[i].kind) {
		return nil, unexpectedToken{tokens[i], i}
	}

	if tokens[i].kind > operator {
		return createBooleaASTWithLeftSideAndOperator(logicalNode{left, right, operator}, tokens[i].kind, tokens[i+1:])
	}

	rightMost, err := createBooleanAST(tokens[i+1:])

	return logicalNode{
		left,
		logicalNode{
			right,
			rightMost,
			tokens[i].kind,
		},
		operator,
	}, err
}

func createBooleanAST(tokens []token) (boolNode, error) {
	left, i, err := boolNodeStartingAt(tokens, 0)
	if err != nil {
		return left, err
	}

	if i >= len(tokens) {
		return left, eoi{}
	}

	if !isOperator(tokens[i].kind) {
		return left, unexpectedToken{tokens[i], i}
	}

	return createBooleaASTWithLeftSideAndOperator(left, tokens[i].kind, tokens[i+1:])
}

func isOperator(kind uint) bool {
	return kind == and || kind == or
}
