package nix

// DefaultPackages which are installed with the rest of nix packages
func DefaultPackages(nixpkgs string) []Dependency {
	return []Dependency{
		{Name: "bash", Nixpkgs: nixpkgs},
		{Name: "coreutils", Nixpkgs: nixpkgs},
		{Name: "gnused", Nixpkgs: nixpkgs},
		{Name: "findutils", Nixpkgs: nixpkgs},
	}
}
