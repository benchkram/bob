package bob

import (
	"path/filepath"

	"github.com/benchkram/bob/bob/global"
)

func (b B) WorkspaceFilePath() string {
	return filepath.Join(b.dir, global.BobWorkspaceFile)
}
