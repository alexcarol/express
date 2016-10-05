package express

import (
	"fmt"
	"strings"
)

const (
	lTrue = iota
	lFalse
	and
)

type token struct {
	kind uint
}

// BoolEval returns the result for an expression
func BoolEval(expression string, variables map[string]interface{}) (bool, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return false, err
	}

	ast, err := createBooleanAST(tokens)
	if err != nil {
		return false, err
	}

	return ast.Eval(variables), nil
}

type boolNode interface {
	Eval(map[string]interface{}) bool
}

type lBool struct {
	value bool
}

func (l lBool) Eval(map[string]interface{}) bool {
	return l.value
}

type logicalNode struct {
	leftBool, rightBool boolNode
	op                  uint
}

func (l logicalNode) Eval(params map[string]interface{}) bool {
	// TODO avoid this switch on expression runtime, consider creating interface for EvalLogical()
	switch l.op {
	case and:
		return l.leftBool.Eval(params) && l.rightBool.Eval(params)
	default:
		// TODO refactor to avoid panics, ASTs should be robust
		panic(fmt.Errorf("unrecognised operator %d", l.op))
	}
}

type eoi struct{}

func (e eoi) Error() string {
	return "End of input"
}

// bool = lTrue|lFalse|bool logicalOperator bool
// logicalOperator = and
func createBooleanAST(tokens []token) (boolNode, error) {
	if len(tokens) == 0 {
		return nil, eoi{}
	}

	switch tokens[0].kind {
	case lTrue, lFalse:
		node := lBool{tokens[0].kind == lTrue}
		if len(tokens) == 1 {
			return node, nil
		}

		if tokens[1].kind == and {

			rightBool, err := createBooleanAST(tokens[2:])
			if err != nil {
				return nil, err
			}

			return logicalNode{node, rightBool, and}, nil
		}
	}

	// TODO return type "text" instead
	return nil, fmt.Errorf("Unexpected token of kind %d", tokens[0].kind)
}

func tokenize(expression string) ([]token, error) {
	var tokens []token

	for i := 0; i < len(expression); i++ {
		if expression[i] == byte(' ') { // TODO add other blank spaces
			continue
		}

		// TOOD find all the "reserved words" generically
		if strings.HasPrefix(expression[i:], "true") {
			i += len("true") - 1
			tokens = append(tokens, token{lTrue})
		} else if strings.HasPrefix(expression[i:], "false") {
			i += len("false") - 1
			tokens = append(tokens, token{lFalse})
		} else if strings.HasPrefix(expression[i:], "and") {
			i += len("and") - 1
			tokens = append(tokens, token{and})
		} else {
			return tokens, fmt.Errorf("Error parsing expression at position %d", i)
		}
	}

	return tokens, nil
}
