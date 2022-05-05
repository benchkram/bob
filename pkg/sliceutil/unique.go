package sliceutil

// Unique remove dublicate entries from the slice.
func Unique(s []string) []string {
	added := make(map[string]bool)
	var res []string
	for _, v := range s {
		if _, exists := added[v]; !exists {
			res = append(res, v)
			added[v] = true
		}
	}
	return res
}
