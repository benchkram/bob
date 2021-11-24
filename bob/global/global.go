package global

import "path/filepath"

const (
	BobFileName      = "bob.yaml"
	BuildToolDir     = ".bob"
	ConfigFile       = "config"
	DefaultBuildTask = "build"
)

// Cache directory
const BobCacheDir = ".bobcache"

var (
	BobCacheBuildinfoDir       = filepath.Join(BobCacheDir, "buildinfos")
	BobCacheTaskHashesFileName = filepath.Join(BobCacheDir, "hashes")
	BobCacheArtifactsDir       = filepath.Join(BobCacheDir, "artifacts")
)
