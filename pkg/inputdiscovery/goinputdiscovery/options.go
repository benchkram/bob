package goinputdiscovery

func WithProjectDir(projectDir string) Option {
	return func(goId *goInputDiscovery) {
		goId.projectDir = projectDir
	}
}
