package express

import "fmt"

const (
	lTrue uint = iota
	lFalse
	and
	or
	variable
)

type token struct {
	kind uint
	text string
}

// BoolEval returns the result for an expression
// TODO think about whether variables should be a single map[string]interface{} or two separate map[string]bool and map[string]<numeric>
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
	case or:
		left, err := l.leftBool.Eval(params)
		if left {
			return true, nil
		}

		// TODO consider ignoring errors but logging them
		// or even making it configurable
		if err != nil {
			return false, err
		}

		return l.rightBool.Eval(params)
	default:
		// TODO refactor to avoid panics, logicalNode should be robust
		// The best is probably to make an "andNode", orNode, ...
		return false, fmt.Errorf("unrecognised operator %d", l.op)
	}
}

func createBooleanAST(tokens []token) (boolNode, error) {
	for _, operator := range []uint{or, and} { // binary logical operators by inverse order of priority
		for i, t := range tokens {
			if t.kind == operator {
				leftNode, err := createBooleanAST(tokens[:i])
				if err != nil {
					return nil, err
				}
				rightNode, err := createBooleanAST(tokens[i+1:])
				if err != nil {
					return nil, err
				}
				return logicalNode{
					leftNode,
					rightNode,
					operator,
				}, nil
			}
		}
	}

	// if we reach this point we have either a variable or a literal or unary operators

	if len(tokens) != 1 {
		return nil, fmt.Errorf("expected tokens length is 1, obtained %d", len(tokens))
	}

	switch tokens[0].kind {
	case lTrue, lFalse:
		return lBool{tokens[0].kind == lTrue}, nil
	case variable:
		return varNode{tokens[0].text}, nil
	default:
		return nil, fmt.Errorf("unexpected token (%s) of kind %d", tokens[0].text, tokens[0].kind)
	}
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
		case "or":
			tokens = append(tokens, token{or, text})
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
