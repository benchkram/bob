package nix

func DefaultPackages() []string {
	return []string{
		"bash",
		"coreutils",
		"gnused",
		"findutils",
	}
}
