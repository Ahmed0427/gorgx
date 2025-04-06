package main

import (
	"testing"
)

func validateEmail(email string) bool {
	regex := `[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+(.[a-zA-Z]{2,})+`
	tokens := parse(regex)
    nfa := toNFA(tokens)
	return match(nfa, email, 0)
}

func TestRegexEngine(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
        // valid
		{"test.email@example.com", true},
		{"test_email123@domain.co", true},
		{"john.doe123@company.io", true},
		{"first_last@domain.name", true},
		{"test.1_2_3@sub.domain.com", true},

        // invalid
		{"test.email@example.com", true},
		{"@example.com", false},         
		{"test@.com", false},           
		{"test@domain", false},        
		{"test@domain..com", false},  
		{"test@domain.c", false},    
		{"@domain.com", false},     
		{"test@domain.123", false},
		{"test@domain.-com", false},      
		{"test@domain,com", false},      
		{" test@email.com", false},     
		{"test@email.com ", false},    
		{" test@domain.com", false},  
		{"test@domain.com ", false},      
	}

	for _, test := range tests {
		t.Run(test.email, func(t *testing.T) {
			result := validateEmail(test.email)
			if result != test.expected {
				t.Errorf("For email %s, expected %v, but got %v",
                    test.email, test.expected, result)
			}
		})
	}
}
