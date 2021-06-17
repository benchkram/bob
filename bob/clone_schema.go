package bob

type CloneSchema string

func (cs *CloneSchema) String() string {
	return string(*cs)
}

const (
	SSH   CloneSchema = "ssh"
	HTTPS CloneSchema = "https"
)
