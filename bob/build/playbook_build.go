package build

import "time"

func (b *Bobfile) BuildPlaybook(taskname string) (*Playbook, error) {
	playbook := NewPlaybook(taskname)
	err := b.Tasks.walk(taskname, "", func(tn string, task Task, err error) error {
		if err != nil {
			return err
		}

		playbook.tasks[tn] = &TaskStatus{
			Task:  task,
			state: StatePending,
			Start: time.Now(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return playbook, nil
}
