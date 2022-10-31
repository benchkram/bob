package bobtask

import (
	"path/filepath"

	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/store"
	"github.com/logrusorgru/aurora"
)

func (t *Task) Dir() string {
	return t.dir
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) ShortName() string {
	_, name := filepath.Split(t.name)
	return name
}

func (t *Task) SetColor(color aurora.Color) {
	t.color = color
}

func (t *Task) ColoredName() string {
	return aurora.Colorize(t.Name(), t.color).String()
}

func (t *Task) Env() []string {
	return t.env
}

func (t *Task) Rebuild() RebuildType {
	return t.rebuild
}

func (t *Task) SetDir(dir string) {
	t.dir = dir
}

func (t *Task) SetName(name string) {
	t.name = name
}

func (t *Task) SetEnv(env []string) {
	t.env = env
}

func (t *Task) Dependencies() []nix.Dependency {
	return t.dependencies
}
func (t *Task) SetDependencies(dependencies []nix.Dependency) {
	t.dependencies = dependencies
}

func (t *Task) SetStorePaths(storePaths []string) {
	t.storePaths = storePaths
}

func (t *Task) StorePaths() []string {
	return t.storePaths
}

func (t *Task) SetNixpkgs(nixpkgs string) {
	t.nixpkgs = nixpkgs
}

func (t *Task) Nixpkgs() string {
	return t.nixpkgs
}

// Project returns the projectname. In case of a non existing projectname the
// tasks local directory is returned.
func (t *Task) Project() string {
	if t.project == "" {
		return t.dir
	}
	return t.project
}

func (t *Task) SetProject(proj string) {
	t.project = proj
}

// Set the rebuild strategy for the task
// defaults to `on-change`.
func (t *Task) SetRebuildStrategy(rebuild RebuildType) {
	t.rebuild = rebuild
}

func (t *Task) WithLocalstore(s store.Store) *Task {
	t.local = s
	return t
}

func (t *Task) WithRemotestore(s store.Store) *Task {
	t.remote = s
	return t
}

func (t *Task) WithBuildinfoStore(s buildinfostore.Store) *Task {
	t.buildInfoStore = s

	return t
}

func (t *Task) WithDockerRegistryClient(c dockermobyutil.RegistryClient) *Task {
	t.dockerRegistryClient = c
	return t
}
