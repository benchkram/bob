package bobrun

const (
	RunTypeCompose RunType = "compose"
	RunTypeBinary  RunType = "binary"
)

type RunType string

func (rt *RunType) String() string {
	return string(*rt)
}
