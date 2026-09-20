package barcodes

import (
	"fmt"
	"strings"
	"unicode"
)

func NormalizeBarcode(raw string) (string, error) {
	var builder strings.Builder
	for _, character := range raw {
		if unicode.IsSpace(character) || character == '-' {
			continue
		}
		if character < '0' || character > '9' {
			return "", fmt.Errorf("barcode must contain digits only")
		}
		builder.WriteRune(character)
	}

	barcode := builder.String()
	if len(barcode) < 8 || len(barcode) > 14 {
		return "", fmt.Errorf("barcode length is invalid")
	}
	return barcode, nil
}
