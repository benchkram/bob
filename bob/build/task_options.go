package build

type TaskOption func(t *Task)

func WithEnvironment(envs []string) TaskOption {
	return func(t *Task) {
		t.env = append(t.env, envs...)
	}
}
