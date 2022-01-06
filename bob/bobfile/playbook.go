package bobfile

import (
	"github.com/Benchkram/bob/bob/playbook"
	"github.com/Benchkram/bob/bobtask"
)

func (b *Bobfile) Playbook(taskname string, opts ...playbook.Option) (*playbook.Playbook, error) {
	pb := playbook.New(
		taskname,
		opts...,
	)

	err := b.Tasks.Walk(taskname, "", func(tn string, task bobtask.Task, err error) error {
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
