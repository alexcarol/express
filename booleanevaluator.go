package express

import "fmt"

const (
	lTrue = iota
	lFalse
	and
	variable
)

type token struct {
	kind uint
	text string
}

// BoolEval returns the result for an expression
func BoolEval(expression string, variables map[string]interface{}) (b bool, err error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return false, err
	}

	ast, err := createBooleanAST(tokens)
	if err != nil {
		return false, err
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
	default:
		// TODO refactor to avoid panics, logicalNode should be robust
		// The best is probably to make an "andNode", orNode, ...
		return false, fmt.Errorf("unrecognised operator %d", l.op)
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

	var node boolNode
	switch tokens[0].kind {
	case lTrue, lFalse:
		node = lBool{tokens[0].kind == lTrue}
	case variable:
		node = varNode{tokens[0].text}
	default:
		return nil, fmt.Errorf("unexpected token (%s) of kind %d", tokens[0].text, tokens[0].kind)
	}

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

	return nil, fmt.Errorf("unexpected token (%s) of kind %d", tokens[0].text, tokens[0].kind)
}

func tokenize(expression string) ([]token, error) {
	var tokens []token

	for i := 0; i < len(expression); i++ {
		if expression[i] == byte(' ') { // TODO add other blank spaces
			continue
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

		text := expression[startingI:i]
		switch text {
		case "true":
			tokens = append(tokens, token{lTrue, text})
		case "false":
			tokens = append(tokens, token{lFalse, text})
		case "and":
			tokens = append(tokens, token{and, text})
		default:
			tokens = append(tokens, token{variable, text})
		}
	}

	return tokens, nil
}

func isValidStarterIdent(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'z'
}

func canBeIdent(b byte) bool {
	return isValidStarterIdent(b)
}
