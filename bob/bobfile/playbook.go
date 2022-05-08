package bobfile

import (
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask"
)

func (b *Bobfile) Playbook(taskName string, opts ...playbook.Option) (*playbook.Playbook, error) {
	pb := playbook.New(
		taskName,
		opts...,
	)

	err := b.BTasks.Walk(taskName, "", func(tn string, task bobtask.Task, err error) error {
		if err != nil {
			return err
		}

		pb.Tasks[tn] = playbook.NewStatus(task)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pb, nil
}
