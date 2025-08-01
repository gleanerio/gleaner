package main

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
)

func main() {
	// Test cases for safeEncodeString function
	testCases := []struct {
		input    string
		expected string
		name     string
	}{
		{
			input:    "https://example.com/path/with[brackets]",
			expected: "path_with_brackets_",
			name:     "URL with brackets",
		},
		{
			input:    "doi:10.1575/path{with}braces",
			expected: "doi:10.1575_path_with_braces",
			name:     "DOI with braces",
		},
		{
			input:    "identifier(with)parentheses",
			expected: "dentifier_with_parentheses",
			name:     "Identifier with parentheses",
		},
		{
			input:    "ark:/12345/path/with/slashes",
			expected: "ark_12345_path_with_slashes",
			name:     "ARK with slashes",
		},
		{
			input:    "mixed/{chars}[test](example)/end",
			expected: "ixed__chars__test__example__end",
			name:     "Mixed problematic characters",
		},
		{
			input:    "simple/string{with}[various](chars)",
			expected: "imple_string_with__various__chars_",
			name:     "Simple string with problematic characters",
		},
	}

	fmt.Println("Testing safeEncodeString function:")
	fmt.Println("====================================")

	for _, test := range testCases {
		result := common.SafeEncodeStringPublic(test.input)
		status := "PASS"
		if result != test.expected {
			status = "FAIL"
		}
		fmt.Printf("Test: %s\n", test.name)
		fmt.Printf("Input:    %s\n", test.input)
		fmt.Printf("Expected: %s\n", test.expected)
		fmt.Printf("Got:      %s\n", result)
		fmt.Printf("Status:   %s\n", status)
		fmt.Println("----")
	}
}
