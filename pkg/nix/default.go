package nix

// DefaultPackages which are installed with the rest of nix packages
func DefaultPackages() []string {
	return []string{
		"bash",
		"coreutils",
		"gnused",
		"findutils",
	}
}
