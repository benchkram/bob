package global

import "path/filepath"

const (
	BobFileName      = "bob.yaml"
	BobWorkspaceFile = ".bob.workspace"

	DefaultBuildTask = "build"
)

// Cache directory
const BobCacheDir = ".bobcache"

var (
	BobCacheBuildinfoDir       = filepath.Join(BobCacheDir, "buildinfos")
	BobCacheTaskHashesFileName = filepath.Join(BobCacheDir, "hashes")
	BobCacheArtifactsDir       = filepath.Join(BobCacheDir, "artifacts")

	BobCacheNix = filepath.Join(BobCacheDir, ".nix_cache")
)
