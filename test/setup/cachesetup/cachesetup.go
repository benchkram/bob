package cachesetup

import (
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bob/global"
)

// Setup creates cache directory structure to be used in a test setup.
func Setup(dir string) (cache string, artifact string, buildInfo string, err error) {
	cache = filepath.Join(dir, global.BobCacheDir)
	err = os.MkdirAll(cache, 0700)
	if err != nil {
		return "", "", "", err
	}
	artifact = filepath.Join(dir, global.BobCacheArtifactsDir)
	err = os.MkdirAll(artifact, 0700)
	if err != nil {
		return "", "", "", err
	}

	buildInfo = filepath.Join(dir, global.BobCacheBuildinfoDir)
	err = os.MkdirAll(buildInfo, 0700)
	if err != nil {
		return "", "", "", err
	}
	return cache, artifact, buildInfo, nil
}
