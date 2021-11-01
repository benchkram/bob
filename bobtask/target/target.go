package target

type Target interface {
	Hash() (string, error)
	Verify() bool
	Exists() bool
	Valid() bool

	WithHash(string) Target
	WithDir(string) Target
}

type T struct {
	// working dir of target
	dir string

	// last computed hash of target
	hash string

	Paths []string   `yaml:"Paths"`
	Type  TargetType `yaml:"Type"`
}

func Make() T {
	return T{
		Paths: []string{},
		Type:  File,
	}
}

func New() *T {
	return new()
}

func new() *T {
	return &T{
		Paths: []string{},
		Type:  File,
	}
}

type TargetType string

const (
	File   TargetType = "file"
	Docker TargetType = "docker"
)

func (t *T) clone() *T {
	target := new()
	target.dir = t.dir
	target.Paths = t.Paths
	target.Type = t.Type
	return target
}

func (t *T) WithDir(dir string) Target {
	target := t.clone()
	target.dir = dir
	return target
}
func (t *T) WithHash(hash string) Target {
	target := t.clone()
	target.hash = hash
	return target
}
