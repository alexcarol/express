package express_test

import (
	"testing"

	"github.com/alexcarol/express"
	"github.com/stretchr/testify/assert"
)

func TestNumericExpressionAllLiterals(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
	}{
		{"5", 5},
		{"55", 55},
		{"((55))", 55},
	}
	for _, testCase := range tests {
		got, err := express.NumericEval(testCase.expression, nil)
		e := "Expression: " + testCase.expression
		assert.NoError(t, err, e)
		assert.Equal(t, testCase.expected, got, e)
	}
}
