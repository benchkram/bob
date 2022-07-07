package bob

import (
	"github.com/benchkram/bob/pkg/boblog"
	"path/filepath"

	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
)

func (b *B) Init() (err error) {
	//b := new()
	return b.init()
}

func (b *B) init() (err error) {
	defer errz.Recover(&err)

	dir := filepath.Join(b.dir, global.BobWorkspaceFile)

	if file.Exists(dir) {
		return usererror.Wrap(ErrWorkspaceAlreadyInitialised)
	}

	boblog.Log.V(1).Info("Initialize Bob workspace...")

	err = b.write()
	errz.Fatal(err)

	return nil
}
