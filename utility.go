package main

import (
	"errors"
	"strconv"
	"strings"
)

func PascalifyShishkebabCase(str string) string {
	return strings.ReplaceAll(strings.Title(strings.ReplaceAll(strings.ToLower(str), "-", " ")), " ", "-")
}

// Parses version string, e.g. "http/1.1", "HTTP/2" or similar
func ParseHttpVersion(versionString string) (major int, minor int, err error) {
	parts := strings.Split(strings.ToLower(versionString), "/")

	if len(parts) != 2 || parts[0] != "http" {
		return 0, 0, errors.New("invalid version format")
	}

	majorMinor := strings.Split(parts[1], ".")

	majorValue, err := strconv.Atoi(majorMinor[0])

	if err != nil {
		return 0, 0, errors.New("invalid version format")
	}

	// Only major
	if len(majorMinor) == 1 {
		return majorValue, 0, nil
	}

	minorValue, err := strconv.Atoi(majorMinor[1])

	if err != nil {
		return 0, 0, errors.New("invalid version format")
	}

	return majorValue, minorValue, nil
}
