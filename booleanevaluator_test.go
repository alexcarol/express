package express

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanEvalWithExpectedErrors(t *testing.T) {
	tests := []struct {
		expression string
	}{
		{"()"},
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, nil)
		assert.Error(t, err)
		assert.False(t, got, "Expression: "+testCase.expression)
	}
}

func TestBooleanEvalAllLiterals(t *testing.T) {
	tests := []struct {
		expression string
		expected   bool
	}{
		{"true", true},
		{"(true)", true},
		{"false", false},
		{"(false)", false},
		{"true or true", true},
		{"true or false", true},
		{"false or true", true},
		{"false or false", false},
		{"true and true", true},
		{"true and false", false},
		{"false and true", false},
		{"false and false", false},
		{"true and true and true", true},
		{"true and true and false", false},
		{"false and true", false},
		{"false and false", false},
		{"false and false and true and true", false},
		{"false and true or true", true},
		{"false and true or false", false},
		{"false or true and true", true},
		{"false or true and false", false},
		{"true and (false or true) and true", true},
		{"false and (true or false) and true", false},
		{"not true", false},
		{"not false", true},
		{"true or not false", true},
		{"true and not false", true},
		{"not true and not false", false},
		{"not true or not true or false", false},
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, nil)
		e := "Expression: " + testCase.expression
		assert.NoError(t, err, e)
		assert.Equal(t, testCase.expected, got, e)
	}
}

func TestBooleanEvalWithVariables(t *testing.T) {
	tests := []struct {
		expression string
		parameters map[string]interface{}
		expected   bool
	}{
		{"potato", map[string]interface{}{"potato": true}, true},
		{"potato", map[string]interface{}{"potato": false}, false},
		{"potato and tomato", map[string]interface{}{"potato": true, "tomato": true}, true},
		{"potato and tomato", map[string]interface{}{"potato": true, "tomato": false}, false},
		{"true and potato and tomato", map[string]interface{}{"potato": true, "tomato": true}, true},
		{"potato and false and tomato", map[string]interface{}{"potato": true, "tomato": true}, false},
		{"potato and (false or tomato)", map[string]interface{}{"potato": true, "tomato": true}, true},
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, testCase.parameters)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expected, got, "Expression: "+testCase.expression)
	}
}

func TestBooleanEvalWithMissingVariables(t *testing.T) {
	tests := []struct {
		expression string
		parameters map[string]interface{}
	}{
		{"potato", nil},
		{"potato and tomato", map[string]interface{}{"potato": true}},
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, testCase.parameters)
		assert.Error(t, err)
		assert.False(t, got, "Expression: "+testCase.expression)
	}
}
