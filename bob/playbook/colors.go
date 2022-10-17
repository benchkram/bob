package playbook

import (
	"fmt"
	"sort"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
)

//	     Run this to pick a new color
//		 	for i := uint8(16); i <= 231; i++ {
//				fmt.Println(i, aurora.Index(i, "pew-pew"))
//			}
var colorPool = []aurora.Color{
	aurora.Gray(18, nil).Color(),
	aurora.GreenFg,
	aurora.BlueFg,
	aurora.CyanFg,
	aurora.MagentaFg,
	aurora.YellowFg,
	aurora.RedFg,
	aurora.Index(42, nil).Color(),
	aurora.Index(45, nil).Color(),
	aurora.Index(51, nil).Color(),
	aurora.Index(111, nil).Color(),
	aurora.Index(141, nil).Color(),
	aurora.Index(225, nil).Color(),
}

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
