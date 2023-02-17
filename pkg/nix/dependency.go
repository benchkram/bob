package nix

// UniqueDeps removes duplicates from the list by checking against name-nixpkgs key
func UniqueDeps(s []Dependency) []Dependency {
	added := make(map[string]bool)
	var res []Dependency
	for _, v := range s {
		if _, exists := added[v.Name+v.Nixpkgs]; !exists {
			res = append(res, v)
			added[v.Name+v.Nixpkgs] = true
		}
	}
	return res
}

func AppendUnique(a []Dependency, x Dependency) []Dependency {
	for _, y := range a {
		if x.Name+x.Nixpkgs == y.Name+y.Nixpkgs {
			return a
		}
	}
	return append(a, x)
}
