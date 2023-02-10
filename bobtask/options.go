package bobtask

type TaskOption func(t *Task)

func WithEnvironment(envs []string) TaskOption {
	return func(t *Task) {
		t.env = append(t.env, envs...)
	}
}

// func WithEnvId(envs []string) TaskOption {
// 	return func(t *Task) {
// 		t.env = append(t.env, envs...)
// 	}
// }
