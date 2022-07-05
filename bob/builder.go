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
	task       string
	aggregator func() *bobfile.Bobfile
	f          BuildFunc
	nix        *NixBuilder
}

func NewBuilder(task string, aggregator func() *bobfile.Bobfile, f BuildFunc) ctl.Builder {
	builder := &builder{
		task:       task,
		aggregator: aggregator,
		f:          f,
	}
	return builder
}

func (b *builder) Build(ctx context.Context) error {
	ag := b.aggregator()
	return b.f(ctx, b.task, ag, b.nix)
}
