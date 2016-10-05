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
