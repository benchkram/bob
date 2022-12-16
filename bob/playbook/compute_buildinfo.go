package playbook

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// computeBuildinfo for a task.
// Should only be called after processing is done.
func (p *Playbook) computeBuildinfo(taskname string) (_ *buildinfo.I, err error) {
	defer errz.Recover(&err)

	fmt.Println("computeBuildinfo", taskname)

	task, ok := p.Tasks[taskname]
	if !ok {
		return nil, usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}

	hashIn, err := task.HashIn()
	errz.Fatal(err)

	buildInfo, err := task.ReadBuildInfo()
	if err != nil {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			// assure buildinfo is initialized correctly
			buildInfo = buildinfo.New()
		} else {
			errz.Fatal(err)
		}
	}
	buildInfo.Meta.Task = task.Name()
	buildInfo.Meta.InputHash = hashIn.String()

	// Compute buildinfo for the target
	trgt, err := task.Task.Target()
	errz.Fatal(err)
	if trgt != nil {
		bi, err := trgt.BuildInfo()
		if err != nil {
			if errors.Is(err, target.ErrTargetDoesNotExist) {
				return nil, usererror.Wrapm(err,
					fmt.Sprintf("Target does not exist for task [%s].\nDid you define the wrong target?\nDid you forget to create the target at all? \n\n", taskname))
			} else {
				errz.Fatal(err)
			}
		}

		buildInfo.Target = *bi
	}

	return buildInfo, nil
}
