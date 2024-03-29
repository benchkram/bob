package global

import "path/filepath"

const (
	BobFileName      = "bob.yaml"
	BobWorkspaceFile = ".bob.workspace"

	DefaultBuildTask = "build"
)

// Cache directory
const BobCacheDir = ".bobcache"
const BobNixCacheFile = ".nix_cache"

var (
	BobCacheBuildinfoDir       = filepath.Join(BobCacheDir, "buildinfos")
	BobCacheTaskHashesFileName = filepath.Join(BobCacheDir, "hashes")
	BobCacheArtifactsDir       = filepath.Join(BobCacheDir, "artifacts")
	BobAuthStoreDir            = filepath.Join(BobCacheDir, "auth")

	BobCacheNixFileName      = filepath.Join(BobCacheDir, BobNixCacheFile)
	BobCacheNixShellCacheDir = filepath.Join(BobCacheDir, "env")
)
