package global

import "path/filepath"

const (
	BobCacheDir      = ".bobcache"
	BobFileName      = "bob.yaml"
	BuildToolDir     = ".bob"
	ConfigFile       = "config"
	DefaultBuildTask = "build"
)

var (
	FileHashesFileName = filepath.Join(BobCacheDir, "filehashes")
	TaskHashesFileName = filepath.Join(BobCacheDir, "hashes")
)
