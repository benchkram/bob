package bob

import (
	"context"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/ctl"
)

var _ ctl.Builder = (*builder)(nil)

type BuildFunc func(_ context.Context, runname string, aggregate *bobfile.Bobfile, nix *NixBuilder) error

// builder holds all dependencies to build a build task
type builder struct {
	task      string
	aggregate *bobfile.Bobfile
	f         BuildFunc
	nix       *NixBuilder
}

func NewBuilder(b *B, task string, aggregate *bobfile.Bobfile, f BuildFunc) ctl.Builder {
	builder := &builder{
		task:      task,
		aggregate: aggregate,
		f:         f,
	}
	return builder
}

func (b *builder) Build(ctx context.Context) error {
	return b.f(ctx, b.task, b.aggregate, b.nix)
}
