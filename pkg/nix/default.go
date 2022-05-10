package nix

// DefaultPackages which are installed with the rest of nix packages
func DefaultPackages() []Dependency {
	return []Dependency{
		{Name: "bash"},
		{Name: "coreutils"},
		{Name: "gnused"},
		{Name: "findutils"},
	}
}
