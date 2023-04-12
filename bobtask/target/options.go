package target

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
