package execctl

type Option func(c *Cmd)

func WithStorePaths(storePaths []string) Option {
	return func(c *Cmd) {
		c.storePaths = storePaths
	}
}

func WithUseNix(useNix bool) Option {
	return func(c *Cmd) {
		c.useNix = useNix
	}
}

func WithArgs(args ...string) Option {
	return func(c *Cmd) {
		c.args = args
	}
}
