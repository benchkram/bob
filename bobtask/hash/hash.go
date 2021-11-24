package hash

// In represents a hash tracking changes of
// inputs, the environment and the task description itself.
// Every change  in one of those must trigger a task rebuild.
type In string

func (i *In) String() string {
	return string(*i)
}
