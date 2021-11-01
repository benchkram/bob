package target

// Valid determines if a target is valid.
// Mostly used to santize input comming from a bobfile.
func (t *T) Valid() bool {
	switch t.Type {
	case File:
		return t.validateFile()
	case Docker:
		return t.validateDocker()
	default:
		// when no target type is set use a file target.
		return t.validateFile()
	}
}

func (t *T) validateFile() bool {
	return len(t.Paths) > 0
}

// validateDocker TODO: implement me
func (t *T) validateDocker() bool {
	return true
}
