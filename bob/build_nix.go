package bob

import (
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/cache"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
)

// Nix builder acts as a wrapper for github.com/benchkram/bob/pkg/nix package
// and is used for building tasks dependencies
type Nix struct {
	// cache allows caching the dependency to store path
	cache cache.Cache
}

type NixOption func(n *Nix)

func WithNixCache(cache cache.Cache) NixOption {
	return func(n *Nix) {
		n.cache = cache
	}
}

// NewNix instantiates a new Nix builder instance
// By default it uses a file to cache the mapping between dependencies and their store paths
// You can override that with WithNixCache option
func NewNix(opts ...NixOption) (_ *Nix, err error) {
	defer errz.Recover(&err)
	var n Nix

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&n)
	}

	if n.cache == nil {
		c, err := nix.NewFileCacheStore()
		errz.Fatal(err)
		n.cache = c
	}

	return &n, nil
}

// BuildNixDependenciesInPipeline collects and builds nix-dependencies for a pipeline starting at taskName.
func (n *Nix) BuildNixDependenciesInPipeline(ag *bobfile.Bobfile, taskName string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	tasksInPipeline, err := ag.BTasks.CollectTasksInPipeline(taskName)
	errz.Fatal(err)

	return n.BuildNixDependencies(ag, tasksInPipeline)
}

// BuildNixDependencies builds nix dependencies and prepares the affected tasks
// by setting the store paths on each task in the given aggregate.
func (n *Nix) BuildNixDependencies(ag *bobfile.Bobfile, tasksInPipeline []string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	nixDependencies, err := ag.BTasks.CollectNixDependenciesForTasks(tasksInPipeline)
	errz.Fatal(err)

	if len(nixDependencies) == 0 {
		return nil
	}

	depStorePathMapping, err := nix.BuildDependencies(
		nix.UniqueDeps(append(nix.DefaultPackages(ag.Nixpkgs), nixDependencies...)),
		n.cache,
	)
	errz.Fatal(err)

	// Resolve nix storePaths from dependencies
	// and rewrite the affected tasks.
	for _, name := range tasksInPipeline {
		t := ag.BTasks[name]

		if !t.UseNix() {
			continue
		}

		// construct used dependencies for this task
		deps := nix.DefaultPackages(ag.Nixpkgs)
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
		errz.Fatal(err)

		t.SetStorePaths(storePaths)
		ag.BTasks[name] = t
	}

	return nil
}

// BuildDependencies builds the list of all nix deps
func (n *Nix) BuildDependencies(deps []nix.Dependency) (nix.DependenciesToStorePathMap, error) {
	return nix.BuildDependencies(deps, n.cache)
}
