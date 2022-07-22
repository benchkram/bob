package sliceutil

// Contains checks if a string is in s
func Contains(s []string, a string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}
