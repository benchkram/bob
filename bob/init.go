package bob

import (
	"os"
	"path/filepath"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/pkg/file"
)

func (b *B) Init() (err error) {
	//b := new()
	return b.init()
}

func (b *B) init() (err error) {
	defer errz.Recover(&err)

	dir := filepath.Join(b.dir, global.BuildToolDir)

	if file.Exists(dir) {
		return ErrBuildToolAlreadyInitialised
	}

	err = os.Mkdir(dir, 0755)
	errz.Fatal(err)

	err = b.write()
	errz.Fatal(err)

	return nil
}
