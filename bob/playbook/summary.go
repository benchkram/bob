package playbook

import (
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
	"github.com/hako/durafmt"
	"github.com/logrusorgru/aurora"
)

// summary prints the tasks processing details as a summary of the playbook.
func (p *Playbook) summary(processedTasks []*bobtask.Task) (err error) {
	defer errz.Recover(&err)

	boblog.Log.V(1).Info("")
	boblog.Log.V(1).Info(aurora.Bold("● ● ● ●").BrightGreen().String())

	duration, err := durafmt.ParseString(p.ExecutionTime().String())
	errz.Fatal(err)

	t := fmt.Sprintf("Ran %d tasks in %s", len(processedTasks), duration.LimitFirstN(1).InternationalString())

	boblog.Log.V(1).Info(aurora.Bold(t).BrightGreen().String())
	for _, t := range processedTasks {
		stat, err := p.TaskStatus(t.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		execTime := ""
		status := stat.State()
		if status != StateNoRebuildRequired {
			execTime = fmt.Sprintf("\t(%s)", stat.ExecutionTime().Round(round))
		}

		taskName := t.Name()
		boblog.Log.V(1).Info(fmt.Sprintf("  %-*s\t%s%s", p.namePad, taskName, status.Summary(), execTime))
	}
	boblog.Log.V(1).Info("")
	return nil
}
