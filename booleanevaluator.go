package express

import "fmt"

const (
	lTrue uint = iota
	lFalse
	and
	or
	variable
)

type eoi struct{}

func (e eoi) Error() string {
	return "End of input"
}

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
		if _, ok := err.(eoi); !ok {
			return false, err
		}
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

func boolNodeFromToken(t token) (boolNode, error) {
	switch t.kind {
	case lTrue, lFalse:
		return lBool{t.kind == lTrue}, nil
	case variable:
		return varNode{t.text}, nil
	default:
		return nil, fmt.Errorf("unexpected token (%s) of kind %d", t.text, t.kind)
	}
}

func createBooleanAST(tokens []token) (boolNode, error) {
	if len(tokens) == 0 {
		return nil, eoi{}
	}

	left, err := boolNodeFromToken(tokens[0])
	if err != nil {
		return nil, err
	}

	var i = 1

	for i < len(tokens) {
		var kind = tokens[i].kind

		if !isOperator(kind) {
			return nil, fmt.Errorf("unexpected token (%s) of kind %d", tokens[i].text, tokens[i].kind)
		}

		i++

		if i == len(tokens) {
			return nil, fmt.Errorf("unexpected end of input")
		}

		right, err := boolNodeFromToken(tokens[i])
		if err != nil {
			return nil, err
		}

		i++

		if i == len(tokens) {
			return logicalNode{left, right, kind}, nil
		}

		if !isOperator(tokens[i].kind) {
			return nil, fmt.Errorf("unexpected token (%s) of kind %d", tokens[i].text, tokens[i].kind)
		}

		if tokens[i].kind > kind {
			left = logicalNode{left, right, kind}
		} else {
			rightMost, err := createBooleanAST(tokens[i+1:])

			return logicalNode{
				left,
				logicalNode{
					right,
					rightMost,
					tokens[i].kind,
				},
				kind,
			}, err
		}
	}

	return left, nil
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

func isOperator(kind uint) bool {
	return kind == and || kind == or
}

func isValidStarterIdent(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'z'
}

func canBeIdent(b byte) bool {
	return isValidStarterIdent(b)
}
