package express

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanEvalAllLiterals(t *testing.T) {
	tests := []struct {
		expression string
		expected   bool
	}{
		{"true", true},
		{"false", false},
		{"true and true", true},
		{"true and false", false},
		{"true and true and true", true},
		{"true and true and false", false},
		{"false and true", false},
		{"false and false", false},
		{"false and false and true and true", false},
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, nil)
		assert.NoError(t, err)
		assert.Equal(t, got, testCase.expected, "Expression: "+testCase.expression)
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
	}
	for _, testCase := range tests {
		got, err := BoolEval(testCase.expression, testCase.parameters)
		assert.NoError(t, err)
		assert.Equal(t, got, testCase.expected, "Expression: "+testCase.expression)
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
