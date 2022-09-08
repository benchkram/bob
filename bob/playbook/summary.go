package playbook

import (
	"fmt"
	"time"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
)

// summary prints the tasks processing details as a summary of the playbook.
func (p *Playbook) summary(processedTasks []*bobtask.Task) {

	boblog.Log.V(1).Info("")
	boblog.Log.V(1).Info(aurora.Bold("● ● ● ●").BrightGreen().String())

	t := fmt.Sprintf("Ran %d tasks in %s", len(processedTasks), displayDuration(p.ExecutionTime()))

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
			execTime = fmt.Sprintf("\t(%s)", displayDuration(stat.ExecutionTime()))
		}

		taskName := t.Name()
		boblog.Log.V(1).Info(fmt.Sprintf("  %-*s\t%s%s", p.namePad, taskName, status.Summary(), execTime))
	}
	boblog.Log.V(1).Info("")
}

func displayDuration(d time.Duration) string {
	if d.Minutes() > 1 {
		return fmt.Sprintf("%.1fm", float64(d)/float64(time.Minute))
	}
	if d.Seconds() > 1 {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	}
	return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond)+0.1) // add .1ms so that it never returns 0.0ms
}
