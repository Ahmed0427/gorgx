package main

import (
	"testing"
	"strings"
)

func validate(regex, input string) bool {
	nfa := toNFA(parse(regex))
	return match(nfa, input)
}

func TestRegexEngine(t *testing.T) {
	tests := []struct {
		regex    string
		input    string
		expected bool
	}{
		{
			regex: "([a-zA-Z]+) ([a-zA-Z]+)",
			input: "John Doe",
			expected: true,
		},
		{
			regex: "([a-zA-Z]+) ([a-zA-Z]+)",
			input: "JohnDoe",
			expected: false,
		},
		{
			regex: "[a-zA-Z0-9]{8,}",
			input: "user1234",
			expected: true,
		},
		{
			regex: "[a-zA-Z0-9]{8,}",
			input: "short7",
			expected: false,
		},
		{
			regex: "(([a-z]+)-([0-9]{4})){3}",
			input: "item-2020item-2021item-2022",
			expected: true,
		},
		{
			regex: "(([a-z]+)-([0-9]{4})){3}",
			input: "item-2020item-2021",
			expected: false,
		},
		{
			regex: "(go)+lang",
			input: "gogogolang",
			expected: true,
		},
		{
			regex: "(go)+lang",
			input: "golang",
			expected: true,
		},
		{
			regex: "(go)+lang",
			input: "gogolanggo",
			expected: false,
		},
		{
			regex: "(([a-z]{2,5}[0-9]{1,3})|([A-Z]{3,5}))+",
			input: "abc12XYZabc345ZZZ",
			expected: true,
		},
		{
			regex: "[a-z]{5,10}[0-9]{3,5}",
			input: "abcdefghij12345",
			expected: true,
		},
		{
			regex: "[a-z]{5,10}[0-9]{3,5}",
			input: "abc123",
			expected: false,
		},
		{
			regex: "(ha)+!",
			input: "hahaha!",
			expected: true,
		},
		{
			regex: "(ha)+!",
			input: "ha!",
			expected: true,
		},
		{
			regex: "(ha)+!",
			input: "huh!",
			expected: false,
		},
		{
			regex: "(([a-z]{3})([0-9]{2})){5}",
			input: "abc12def34ghi56jkl78mno90",
			expected: true,
		},
		{
			regex: "(([a-z]{3})([0-9]{2})){5}",
			input: "abc12def34ghi56jkl78",
			expected: false,
		},
		{
			regex: "a*a*a*a*a*a*a*a*a*a*",
			input: "aaaaaaaaaaaaaaaaaaab",
			expected: false,
		},
		{
			regex: "a{100}",
			input: strings.Repeat("a", 100),
			expected: true,
		},
		{
			regex: "a{100}",
			input: strings.Repeat("a", 99),
			expected: false,
		},
		{
			regex: "(ab|cd){50}",
			input: strings.Repeat("ab", 25) + strings.Repeat("cd", 25),
			expected: true,
		},
	}

	for _, test := range tests {
		result := validate(test.regex, test.input)
		if result != test.expected {
			t.Errorf("\nRegex: %s\nInput: %s\nExpected: %v\nOutputed: %v\n",
				test.regex, test.input, test.expected, result)
		}
	}
}
