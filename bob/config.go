package bob

import (
	"path/filepath"

	"github.com/Benchkram/bob/bob/global"
)

func (b B) ConfigFilePath() string {
	return filepath.Join(b.dir, global.BuildToolDir, global.ConfigFile)
}
