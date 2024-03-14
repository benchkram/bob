package nixbuilder

import (
	"fmt"
	"slices"

	"github.com/benchkram/bob/pkg/envutil"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
)

// NB acts as a wrapper for github.com/benchkram/bob/pkg/nix package
// and is used for building tasks dependencies
type NB struct {
	// cache allows caching the dependency to store path
	cache *nix.Cache

	// shellCache allows caching of the nix-shell --command='env' output
	shellCache *nix.ShellCache

	// envStore is filled by NixBuilder with the environment
	// used by tasks.
	envStore envutil.Store
}

type NixOption func(n *NB)

func WithCache(cache *nix.Cache) NixOption {
	return func(n *NB) {
		n.cache = cache
	}
}

func WithShellCache(cache *nix.ShellCache) NixOption {
	return func(n *NB) {
		n.shellCache = cache
	}
}

func WithEnvironmentStore(store envutil.Store) NixOption {
	return func(n *NB) {
		n.envStore = store
	}
}

// NewNB instantiates a new Nix builder instance
func New(opts ...NixOption) *NB {
	n := &NB{
		envStore: envutil.NewStore(),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(n)
	}

	return n
}

func (n *NB) EnvStore() envutil.Store {
	return n.envStore
}

// BuildNixDependenciesInPipeline collects and builds nix-dependencies for a pipeline starting at taskName.
func (n *NB) BuildNixDependenciesInPipeline(ag *bobfile.Bobfile, taskName string) (err error) {
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
func (n *NB) BuildNixDependencies(ag *bobfile.Bobfile, buildTasksInPipeline, runTasksInPipeline []string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	var shellDotNix *string
	var shellDotNixHash *string
	if ag.ShellDotNix != "" {
		shellDotNix = &ag.ShellDotNix

		// When a folder is given instead of the direct path to shell.nix
		// bob assumes that the folder contains a shell.nix file.
		// TODO: Implement me.

		// Concat and sort files to ensure consistent hash
		shellDotNixFiles := []string{ag.ShellDotNix}
		shellDotNixFiles = append(shellDotNixFiles, ag.ShellDotNixImports...)
		slices.Sort(shellDotNixFiles)
		// Copute hash of shell.nix and its imports
		hash, err := filehash.HashOfFiles(shellDotNixFiles...)
		errz.Fatal(err)

		shellDotNixHash = &hash
	}

	// Resolve nix storePaths from dependencies
	// and rewrite the affected tasks.
	for _, name := range buildTasksInPipeline {
		t := ag.BTasks[name]

		// construct used dependencies for this task
		var deps []nix.Dependency
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		t.SetNixpkgs(ag.Nixpkgs)

		hash, err := nix.HashDependencies(deps)
		errz.Fatal(err)

		if _, ok := n.envStore[envutil.Hash(hash)]; !ok {
			nixShellEnv, err := n.BuildEnvironment(deps, ag.Nixpkgs,
				BuildEnvironmentArgs{
					ShellDotNix:     shellDotNix,
					ShellDotNixHash: shellDotNixHash,
				},
			)
			errz.Fatal(err)
			n.envStore[envutil.Hash(hash)] = nixShellEnv
		}
		t.SetEnvID(envutil.Hash(hash))

		ag.BTasks[name] = t
	}

	// FIXME: environment cache is a workaround...
	// either use envSTore and adapt run tasks to use it as well
	// or remove run tasks entirely.
	environmentCache := make(map[string][]string)
	for _, name := range runTasksInPipeline {
		t := ag.RTasks[name]

		// construct used dependencies for this task
		var deps []nix.Dependency
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		t.SetNixpkgs(ag.Nixpkgs)

		hash, err := nix.HashDependencies(deps)
		errz.Fatal(err)

		if _, ok := environmentCache[hash]; !ok {
			nixShellEnv, err := n.BuildEnvironment(deps, ag.Nixpkgs,
				BuildEnvironmentArgs{
					ShellDotNix:     shellDotNix,
					ShellDotNixHash: shellDotNixHash,
				},
			)
			errz.Fatal(err)
			environmentCache[hash] = nixShellEnv
		}
		t.SetEnv(envutil.Merge(environmentCache[hash], t.Env()))

		ag.RTasks[name] = t
	}

	return nil
}

// BuildDependencies builds the list of all nix deps
func (n *NB) BuildDependencies(deps []nix.Dependency) error {
	return nix.BuildDependencies(deps, n.cache)
}

type BuildEnvironmentArgs struct {
	ShellDotNix     *string
	ShellDotNixHash *string
}

// BuildEnvironment builds the environment with all nix deps
func (n *NB) BuildEnvironment(deps []nix.Dependency, nixpkgs string, args BuildEnvironmentArgs) (_ []string, err error) {
	return nix.BuildEnvironment(deps, nixpkgs,
		nix.BuildEnvironmentArgs{
			Cache:           n.cache,
			ShellCache:      n.shellCache,
			ShellDotNix:     args.ShellDotNix,
			ShellDotNixHash: args.ShellDotNixHash,
		},
	)
}

// Clean removes all cached nix dependencies
func (n *NB) Clean() (err error) {
	return n.cache.Clean()
}

// CleanNixShellCache removes all cached nix-shell --command='env' output
func (n *NB) CleanNixShellCache() (err error) {
	if n.shellCache == nil {
		return nil
	}
	return n.shellCache.Clean()
}
