package evaluator

import (
	"fmt"
	"github.com/muiscript/ether/lexer"
	"github.com/muiscript/ether/object"
	"github.com/muiscript/ether/parser"
	"testing"
)

func TestEval_Integer(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		expected int
	}{
		{
			desc:     "42",
			input:    "42;",
			expected: 42,
		},
		{
			desc:     "-42",
			input:    "-42;",
			expected: -42,
		},
		{
			desc:     "-(-42)",
			input:    "-(-42);",
			expected: 42,
		},
		{
			desc:     "15 + 3",
			input:    "15 + 3;",
			expected: 18,
		},
		{
			desc:     "15 - 3",
			input:    "15 - 3;",
			expected: 12,
		},
		{
			desc:     "15 * 3",
			input:    "15 * 3;",
			expected: 45,
		},
		{
			desc:     "15 / 3",
			input:    "15 / 3;",
			expected: 5,
		},
		{
			desc:     "1 + 2 * 3",
			input:    "1 + 2 * 3;",
			expected: 7,
		},
		{
			desc:     "(1 + 2) * 3",
			input:    "(1 + 2) * 3;",
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			evaluated := eval(t, tt.input)
			integer, ok := evaluated.(*object.Integer)
			if !ok {
				t.Errorf("unable to convert to integer: %+v\n", evaluated)
			}
			if integer.Value != tt.expected {
				t.Errorf("integer value wrong.\nwant=%d\ngot=%d\n", tt.expected, integer.Value)
			}
		})
	}
}

// since the parse of function literal is tested in parser package,
// here we only test whether...
// - the function literal is evaluated as function object
// - the environment is properly included in function.
func TestEval_Function(t *testing.T) {
	tests := []struct {
		desc                string
		input               string
		expectedEnvVarName  []string
		expectedEnvVarValue []interface{}
	}{
		{
			desc:                "|| { 42; };",
			input:               "|| { 42; };",
			expectedEnvVarName:  []string{},
			expectedEnvVarValue: []interface{}{},
		},
		{
			desc:                "var c = 1; || { 42; };",
			input:               "var c = 1; || { 42; };",
			expectedEnvVarName:  []string{"c"},
			expectedEnvVarValue: []interface{}{1},
		},
		{
			desc:                "var a = 2; var b = 3; || { 42; };",
			input:               "var a = 2; var b = 3; || { 42; };",
			expectedEnvVarName:  []string{"a", "b"},
			expectedEnvVarValue: []interface{}{2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			evaluated := eval(t, tt.input)
			function, ok := evaluated.(*object.Function)
			fmt.Printf("%+v\n", function)
			fmt.Printf("%+v\n", function.Env)
			if !ok {
				t.Errorf("unable to convert to function: %+v\n", evaluated)
			}
			for i, expectedName := range tt.expectedEnvVarName {
				expectedValue := tt.expectedEnvVarValue[i]
				actual := function.Env.Get(expectedName)
				if actual == nil {
					t.Errorf("undefined identifier: %s\n", expectedName)
				}
				testObject(t, expectedValue, actual)
			}
		})
	}
}

func TestEval_VarStatement(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		expected int
	}{
		{
			desc:     "assignment",
			input:    "var a = 42; a;",
			expected: 42,
		},
		{
			desc:     "operation using identifier",
			input:    "var a = 42; a / 2;",
			expected: 21,
		},
		{
			desc:     "re-assignment",
			input:    "var a = 42; var b = a; b;",
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			evaluated := eval(t, tt.input)
			integer, ok := evaluated.(*object.Integer)
			if !ok {
				t.Errorf("unable to convert to integer: %+v\n", evaluated)
			}
			if integer.Value != tt.expected {
				t.Errorf("integer value wrong.\nwant=%d\ngot=%d\n", tt.expected, integer.Value)
			}
		})
	}
}

func eval(t *testing.T, input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)

	program, err := p.ParseProgram()
	if err != nil {
		t.Errorf("parse error: %s\n", err.Error())
	}

	env := object.NewEnvironment()
	evaluated, err := Eval(program, env)
	if err != nil {
		t.Errorf("eval error: %s\n", err.Error())
	}

	return evaluated
}

func testObject(t *testing.T, expectedValue interface{}, actual object.Object) {
	switch expectedValue := expectedValue.(type) {
	case int:
		if actualValue := actual.(*object.Integer).Value; actualValue != expectedValue {
			t.Errorf("integer value wrong:\nwant=%d\ngot=%d\n", expectedValue, actualValue)
		}
	default:
		t.Errorf("unexpected type: %T", expectedValue)
	}
}
