package bobgit

// Status executes `git add` in all repositories
// first level repositories found inside a .bob filtree.
// It parses the output of each call and creates a object
// containing status infos for all of them combined.
// The result is similar to what `git add` would print
// but visualy optimised for the multi repository case.
func Add(url string) error {
	return nil
}
