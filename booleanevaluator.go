package express

import (
	"errors"
	"fmt"
	"strings"
)

const (
	lTrue uint = iota
	lFalse
	and
	or
	variable
	lParen
	rParen
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

func boolNodeFromToken(tokens []token, position int) (boolNode, error) {
	if len(tokens) <= position {
		return nil, eoi{}
	}
	t := tokens[position]
	switch t.kind {
	case lTrue, lFalse:
		return lBool{t.kind == lTrue}, nil
	case variable:
		return varNode{t.text}, nil
	default:
		return nil, unexpectedToken{t, position}
	}
}

func createBooleaASTWithLeftSideAndOperator(left boolNode, operator uint, tokens []token) (boolNode, error) {
	var i = 0
	right, err := boolNodeFromToken(tokens, i)
	if err != nil {
		utErr, ok := err.(unexpectedToken)
		if !ok || utErr.t.kind != lParen {
			return left, fmt.Errorf("unexpected: %v", err)
		}

		right, err = createBooleanAST(tokens[i+1:])

		utErr, ok = err.(unexpectedToken)
		if !ok || utErr.t.kind != rParen {
			return left, err
		}

		i += utErr.position
	}

	i++

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
	var i = 1

	left, err := boolNodeFromToken(tokens, 0)
	if err != nil {
		utErr, ok := err.(unexpectedToken)
		if !ok || utErr.t.kind != lParen {
			return left, err
		}

		left, err = createBooleanAST(tokens[1:])

		utErr, ok = err.(unexpectedToken)
		if !ok || utErr.t.kind != rParen {
			return left, err
		}

		i = utErr.position + 1
	}

	if i >= len(tokens) {
		return left, eoi{}
	}

	if !isOperator(tokens[i].kind) {
		// TODO see if it makes sense to pass the position, as it is relative
		return left, unexpectedToken{tokens[i], i}
	}

	return createBooleaASTWithLeftSideAndOperator(left, tokens[i].kind, tokens[i+1:])
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

func isOperator(kind uint) bool {
	return kind == and || kind == or
}

func isValidStarterIdent(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'z'
}

func canBeIdent(b byte) bool {
	return isValidStarterIdent(b)
}
