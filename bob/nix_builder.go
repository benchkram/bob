package bob

import (
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
)

// NixBuilder acts as a wrapper for github.com/benchkram/bob/pkg/nix package
// and is used for building tasks dependencies
type NixBuilder struct {
	// cache allows caching the dependency to store path
	cache *nix.Cache
}

type NixOption func(n *NixBuilder)

func WithCache(cache *nix.Cache) NixOption {
	return func(n *NixBuilder) {
		n.cache = cache
	}
}

// NewNixBuilder instantiates a new Nix builder instance
func NewNixBuilder(opts ...NixOption) *NixBuilder {
	n := &NixBuilder{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(n)
	}

	return n
}

// BuildNixDependenciesInPipeline collects and builds nix-dependencies for a pipeline starting at taskName.
func (n *NixBuilder) BuildNixDependenciesInPipeline(ag *bobfile.Bobfile, taskName string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	tasksInPipeline, err := ag.BTasks.CollectTasksInPipeline(taskName)
	errz.Fatal(err)

	return n.BuildNixDependencies(ag, tasksInPipeline, []string{})
}

// BuildNixDependencies builds nix dependencies and prepares the affected tasks
// by setting the store paths on each task in the given aggregate.
func (n *NixBuilder) BuildNixDependencies(ag *bobfile.Bobfile, buildTasksInPipeline, runTasksInPipeline []string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	nixDependencies, err := ag.BTasks.CollectNixDependenciesForTasks(buildTasksInPipeline)
	errz.Fatal(err)

	runTasksDependencies, err := ag.RTasks.CollectNixDependenciesForTasks(runTasksInPipeline)
	errz.Fatal(err)
	nixDependencies = append(nixDependencies, runTasksDependencies...)

	depStorePathMapping, err := nix.BuildDependencies(
		nix.UniqueDeps(nixDependencies),
		n.cache,
	)
	errz.Fatal(err)

	// Resolve nix storePaths from dependencies
	// and rewrite the affected tasks.
	for _, name := range buildTasksInPipeline {
		t := ag.BTasks[name]

		// construct used dependencies for this task
		var deps []nix.Dependency
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
		errz.Fatal(err)

		t.SetStorePaths(storePaths)
		t.SetNixpkgs(ag.Nixpkgs)
		ag.BTasks[name] = t
	}

	for _, name := range runTasksInPipeline {
		t := ag.RTasks[name]

		// construct used dependencies for this task
		var deps []nix.Dependency
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
		errz.Fatal(err)

		t.SetStorePaths(storePaths)
		t.SetNixpkgs(ag.Nixpkgs)

		ag.RTasks[name] = t
	}

	return nil
}

// BuildDependencies builds the list of all nix deps
func (n *NixBuilder) BuildDependencies(deps []nix.Dependency) (nix.DependenciesToStorePathMap, error) {
	return nix.BuildDependencies(deps, n.cache)
}
