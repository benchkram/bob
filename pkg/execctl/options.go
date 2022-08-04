package execctl

type Option func(c *Cmd)

func WithStorePaths(storePaths []string) Option {
	return func(c *Cmd) {
		c.storePaths = storePaths
	}
}

func WithArgs(args ...string) Option {
	return func(c *Cmd) {
		c.args = args
	}
}

func WithEnv(env []string) Option {
	return func(c *Cmd) {
		c.env = env
	}
}
