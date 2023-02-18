package main

import (
	"strings"
)

func PascalifyShishkebabCase(str string) string {
	return strings.ReplaceAll(strings.Title(strings.ReplaceAll(strings.ToLower(str), "-", " ")), " ", "-")
}
