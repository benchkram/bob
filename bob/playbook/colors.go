package playbook

import (
	"fmt"
	"sort"
	"time"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
)

var colorPool = []aurora.Color{
	1,
	aurora.BlueFg,
	aurora.GreenFg,
	aurora.CyanFg,
	aurora.MagentaFg,
	aurora.YellowFg,
	aurora.RedFg,
}
var round = 10 * time.Millisecond

// pickTaskColors picks a display color for each task in the playbook.
func (p *Playbook) pickTaskColors() {
	tasks := []string{}
	for _, t := range p.Tasks {
		tasks = append(tasks, t.Name())
	}
	sort.Strings(tasks)

	// Adjust padding of first column based on the taskname length.
	// Also assign fixed color to the tasks.
	p.namePad = 0
	for i, name := range tasks {
		if len(name) > p.namePad {
			p.namePad = len(name)
		}

		color := colorPool[i%len(colorPool)]
		p.Tasks[name].Task.SetColor(color)
	}
	p.namePad += 14

	dependencies := len(tasks) - 1
	rootName := p.Tasks[p.root].ColoredName()
	boblog.Log.V(1).Info(fmt.Sprintf("Running task %s with %d dependencies", rootName, dependencies))
}
