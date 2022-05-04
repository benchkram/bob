package playbook

type Option func(p *Playbook)

func WithCachingEnabled(enable bool) Option {
	return func(p *Playbook) {
		p.enableCaching = enable
	}
}

func WithPkgToStorePath(pkgToStorePath map[string]string) Option {
	return func(p *Playbook) {
		p.pkgToStorePath = pkgToStorePath
	}
}
