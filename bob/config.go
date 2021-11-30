package bob

import (
	"path/filepath"

	"github.com/Benchkram/bob/bob/global"
)

func (b B) WorkspaceFilePath() string {
	return filepath.Join(b.dir, global.BobWorkspaceFile)
}
