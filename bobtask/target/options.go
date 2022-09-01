package target

import "github.com/benchkram/bob/bobtask/buildinfo"

type Option func(t *T)

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

func WithExpected(bi *buildinfo.Targets) Option {
	return func(t *T) {
		t.expected = bi
	}
}
