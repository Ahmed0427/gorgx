package main

import (
	"testing"
)

func validate(regex, input string) bool {
	nfa := toNFA(parse(regex))
	return match(nfa, input, 0)
}

func TestRegexEngine(t *testing.T) {
	tests := []struct {
		regex    string
		input    string
		expected bool
	}{
		{
			regex: "([0-9]{4})/([0-9]{2})/([0-9]{2})",
			input: "2025/04/14",
			expected: true,
		},
	}

	for _, test := range tests {
		result := validate(test.regex, test.input)
		if result != test.expected {
			t.Errorf("For email %s, expected %v, but got %v",
				test.input, test.expected, result)
		}
	}
}
