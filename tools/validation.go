package tools

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

const (
	maxDimension        = 1e6 // 1,000,000 mm = 1 km
	maxStringLength     = 1024
	maxObjectNameLength = 256
)

var objectNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_ ]*$`)

// ValidateNumber mirrors validateNumber(value, name, {min, max, allowZero:true}).
func ValidateNumber(value any, name string, min, max float64) (float64, error) {
	num, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("Invalid %s: must be a finite number, got %T", name, value)
	}
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return 0, fmt.Errorf("Invalid %s: must be a finite number, got %v", name, num)
	}
	if num < min {
		return 0, fmt.Errorf("Invalid %s: %v is below minimum %v", name, num, min)
	}
	if num > max {
		return 0, fmt.Errorf("Invalid %s: %v exceeds maximum %v", name, num, max)
	}
	return num, nil
}

// ValidatePositiveNumber mirrors validatePositiveNumber: (> 0, default max 1e6).
func ValidatePositiveNumber(value any, name string, maxes ...float64) (float64, error) {
	max := maxDimension
	if len(maxes) > 0 {
		max = maxes[0]
	}
	return ValidateNumber(value, name, math.Nextafter(0, 1), max)
}

// ValidateNonNegativeNumber mirrors validateNonNegativeNumber (>= 0, default max 1e6).
func ValidateNonNegativeNumber(value any, name string, maxes ...float64) (float64, error) {
	max := maxDimension
	if len(maxes) > 0 {
		max = maxes[0]
	}
	return ValidateNumber(value, name, 0, max)
}

// ValidateString mirrors validateString: type check, length limit.
func ValidateString(value any, name string, maxLengths ...int) (string, error) {
	maxLen := maxStringLength
	if len(maxLengths) > 0 {
		maxLen = maxLengths[0]
	}
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Invalid %s: must be a string, got %T", name, value)
	}
	if len(s) == 0 {
		return "", fmt.Errorf("Invalid %s: must not be empty", name)
	}
	if len(s) > maxLen {
		return "", fmt.Errorf("Invalid %s: exceeds maximum length of %d characters", name, maxLen)
	}
	return s, nil
}

// ValidateObjectName mirrors validateObjectName.
func ValidateObjectName(value any, name string) (string, error) {
	s, err := ValidateString(value, name, maxObjectNameLength)
	if err != nil {
		return "", err
	}
	if !objectNameRe.MatchString(s) {
		return "", fmt.Errorf("Invalid %s: %q contains invalid characters. Use alphanumeric, underscores, and spaces only.", name, s)
	}
	return s, nil
}

// EscapePythonString escapes a string for safe embedding in Python code.
func EscapePythonString(s string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\x00", "",
	)
	return r.Replace(s)
}

// ValidateFilePath mirrors validateFilePath: absolute paths only, traversal
// and system path blocking.
func ValidateFilePath(value any, name string) (string, error) {
	s, err := ValidateString(value, name, 4096)
	if err != nil {
		return "", err
	}
	if strings.Contains(s, "\x00") {
		return "", fmt.Errorf("Invalid %s: path contains null bytes", name)
	}
	isWinAbs := len(s) >= 3 && isLetter(s[0]) && s[1] == ':' && (s[2] == '\\' || s[2] == '/')
	if !strings.HasPrefix(s, "/") && !isWinAbs {
		return "", fmt.Errorf("Invalid %s: must be an absolute path", name)
	}
	normalized := strings.ReplaceAll(s, "\\", "/")
	if strings.Contains(normalized, "/../") || strings.HasSuffix(normalized, "/..") || strings.HasPrefix(normalized, "../") {
		return "", fmt.Errorf("Invalid %s: path traversal detected", name)
	}
	lower := strings.ToLower(normalized)
	for _, prefix := range []string{"/etc/", "/proc/", "/sys/", "/dev/", "/var/run/", "/var/log/"} {
		if strings.HasPrefix(lower, prefix) {
			return "", fmt.Errorf("Invalid %s: access to %s is not allowed", name, prefix)
		}
	}
	return s, nil
}

// ValidateArray mirrors validateArray: must be non-empty array within limits.
func ValidateArray(value any, name string, maxLengths ...int) ([]any, error) {
	maxLen := 10000
	if len(maxLengths) > 0 {
		maxLen = maxLengths[0]
	}
	arr, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("Invalid %s: must be an array, got %T", name, value)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("Invalid %s: must not be empty", name)
	}
	if len(arr) > maxLen {
		return nil, fmt.Errorf("Invalid %s: exceeds maximum length of %d items", name, maxLen)
	}
	return arr, nil
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
