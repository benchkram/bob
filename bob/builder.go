package bob

import (
	"context"

	"github.com/benchkram/bob/bob/bobfile"
	nixbuilder "github.com/benchkram/bob/bob/nix-builder"
	"github.com/benchkram/bob/pkg/ctl"
)

var _ ctl.Builder = (*builder)(nil)

type BuildFunc func(_ context.Context, runname string, aggregate *bobfile.Bobfile, nix *nixbuilder.NB) error

// builder holds all dependencies to build a build task
type builder struct {
	task      string
	aggregate *bobfile.Bobfile
	f         BuildFunc
	nix       *nixbuilder.NB
}

func NewBuilder(task string, aggregate *bobfile.Bobfile, f BuildFunc, nix *nixbuilder.NB) ctl.Builder {
	builder := &builder{
		task:      task,
		aggregate: aggregate,
		f:         f,
		nix:       nix,
	}
	return builder
}

func (b *builder) Build(ctx context.Context) error {
	return b.f(ctx, b.task, b.aggregate, b.nix)
}
