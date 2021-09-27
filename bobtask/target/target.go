package target

type T struct {
	Paths []string
	Type  TargetType
}

func Make() T {
	return T{
		Paths: []string{},
		Type:  File,
	}
}

type TargetType string

const (
	File   TargetType = "file"
	Docker TargetType = "docker"
)
