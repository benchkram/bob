package target

import (
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Option func(t *T)

func WithDir(dir string) Option {
	return func(t *T) {
		t.dir = dir
	}
}

func WithFilesystemEntries(entries []string) Option {
	return func(t *T) {
		t.filesystemEntriesRaw = entries
	}
}

func WithDockerImages(images []string) Option {
	return func(t *T) {
		t.dockerImages = images
	}
}

func WithDockerRegistryClient(dockerRegistryClient dockermobyutil.RegistryClient) Option {
	return func(t *T) {
		t.dockerRegistryClient = dockerRegistryClient
	}
}

func WithExpected(bi *buildinfo.Targets) Option {
	return func(t *T) {
		t.expected = bi
	}
}
