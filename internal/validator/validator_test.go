package validator

import "testing"

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "valid username", input: "john", expected: true},
		{name: "exactly 20 chars", input: "abcdefghijklmnopqrst", expected: true},
		{name: "single char", input: "a", expected: true},
		{name: "empty string", input: "", expected: false},
		{name: "21 chars (too long)", input: "abcdefghijklmnopqrstu", expected: false},
		{name: "contains space", input: "john doe", expected: false},
		{name: "leading space", input: " john", expected: false},
		{name: "trailing space", input: "john ", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUsername(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidUsername(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "valid password", input: "secret123", expected: true},
		{name: "exactly 4 chars (min)", input: "abcd", expected: true},
		{name: "exactly 100 chars (max)", input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expected: true},
		{name: "empty string", input: "", expected: false},
		{name: "3 chars (too short)", input: "abc", expected: false},
		{name: "101 chars (too long)", input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPassword(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidPassword(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
