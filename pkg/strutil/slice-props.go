package strutil

import "strings"

// Contains returns true if slice contains element
func Contains(slice []string, element string) bool {
	return !(PosString(slice, element) == -1)
}

// PosString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func PosString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// LongestStrLen returns maximum string length from a slice of strings
func LongestStrLen(inputs []string) int {
	maxlen := -1

	for _, i := range inputs {
		if len(i) > maxlen {
			maxlen = len(i)
		}
	}

	return maxlen
}

// ConvertToLines converts bytes into a list of strings separeted by newline
func ConvertToLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")
}
