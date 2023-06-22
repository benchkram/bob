package bobfile

import (
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask"
)

func (b *Bobfile) Playbook(taskName string, opts ...playbook.Option) (*playbook.Playbook, error) {

	var idCounter int
	pb := playbook.New(
		taskName,
		idCounter,
		opts...,
	)
	idCounter++

	err := b.BTasks.Walk(taskName, "", func(tn string, task bobtask.Task, err error) error {
		if err != nil {
			return err
		}
		if taskName == tn {
			// The root task already has an id
			statusTask := playbook.NewStatus(&task)
			pb.Tasks[tn] = statusTask
			pb.TasksOptimized = append(pb.TasksOptimized, statusTask)
			return nil
		}

		task.TaskID = idCounter
		statusTask := playbook.NewStatus(&task)

		pb.Tasks[tn] = statusTask
		pb.TasksOptimized = append(pb.TasksOptimized, statusTask)

		idCounter++

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pb, nil
}
