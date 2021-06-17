package bob

import (
	"path/filepath"
)

const (
	BuildToolDir = ".bob"
	ConfigFile   = "config"
)

func (b B) ConfigFilePath() string {
	return filepath.Join(b.dir, BuildToolDir, ConfigFile)
}
