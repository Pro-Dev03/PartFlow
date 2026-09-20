package barcodes

import "testing"

func TestNormalizeBarcode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{name: "spaces and hyphens", input: " 0 360-00291452 ", expected: "036000291452", valid: true},
		{name: "leading zero", input: "0123456789012", expected: "0123456789012", valid: true},
		{name: "letters", input: "ABC-12345678", valid: false},
		{name: "too short", input: "1234567", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := NormalizeBarcode(test.input)
			if test.valid {
				if err != nil || actual != test.expected {
					t.Fatalf("NormalizeBarcode(%q) = %q, %v; want %q", test.input, actual, err, test.expected)
				}
				return
			}
			if err == nil {
				t.Fatalf("NormalizeBarcode(%q) unexpectedly succeeded with %q", test.input, actual)
			}
		})
	}
}
