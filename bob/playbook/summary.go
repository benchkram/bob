package playbook

import (
	"fmt"

	"github.com/benchkram/bob/bobtask/processed"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/format"
	"github.com/logrusorgru/aurora"
)

// summary prints the tasks processing details as a summary of the playbook.
func (p *Playbook) summary(processedTasks []*processed.Task) {

	boblog.Log.V(1).Info("")
	boblog.Log.V(1).Info(aurora.Bold("● ● ● ●").BrightGreen().String())

	t := fmt.Sprintf("Ran %d tasks in %s", len(processedTasks), format.DisplayDuration(p.ExecutionTime()))

	boblog.Log.V(1).Info(aurora.Bold(t).BrightGreen().String())
	for _, t := range processedTasks {
		stat, err := p.TaskStatus(t.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		execTime := ""
		status := stat.State()
		//if status != StateNoRebuildRequired {
		execTime = fmt.Sprintf("\t(%s)", format.DisplayDuration(stat.ExecutionTime()))
		//}

		taskName := t.Name()
		boblog.Log.V(1).Info(fmt.Sprintf("  %-*s\t%s%s", p.namePad, taskName, status.Summary(), execTime))

	}
	boblog.Log.V(1).Info("")
}
