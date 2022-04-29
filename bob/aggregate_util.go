package bob

import (
	"strings"

	"github.com/benchkram/bob/bob/bobfile"
)

// syncProjectName project names for all bobfiles and build tasks
func syncProjectName(
	a *bobfile.Bobfile,
	bobs []*bobfile.Bobfile,
) (*bobfile.Bobfile, []*bobfile.Bobfile) {
	for _, bobfile := range bobs {
		bobfile.Project = a.Project

		for taskname, task := range bobfile.BTasks {
			// Should be the name of the umbrella-bobfile.
			task.SetProject(a.Project)

			// Overwrite value in build map
			bobfile.BTasks[taskname] = task
		}
	}

	return a, bobs
}

func (b *B) addBuildTasksToAggregate(
	a *bobfile.Bobfile,
	bobs []*bobfile.Bobfile,
) *bobfile.Bobfile {

	for _, bobfile := range bobs {
		// Skip the aggregate
		if bobfile.Dir() == a.Dir() {
			continue
		}

		for taskname, task := range bobfile.BTasks {
			dir := bobfile.Dir()

			// Use a relative path as task prefix.
			prefix := strings.TrimPrefix(dir, b.dir)
			taskname := addTaskPrefix(prefix, taskname)

			// Alter the taskname.
			task.SetName(taskname)

			// Rewrite dependent tasks to global scope.
			dependsOn := []string{}
			for _, dependentTask := range task.DependsOn {
				dependsOn = append(dependsOn, addTaskPrefix(prefix, dependentTask))
			}
			task.DependsOn = dependsOn

			a.BTasks[taskname] = task
		}
	}

	return a
}

func (b *B) addRunTasksToAggregate(
	a *bobfile.Bobfile,
	bobs []*bobfile.Bobfile,
) *bobfile.Bobfile {

	for _, bobfile := range bobs {
		// Skip the aggregate
		if bobfile.Dir() == a.Dir() {
			continue
		}

		for runname, run := range bobfile.RTasks {
			dir := bobfile.Dir()

			// Use a relative path as task prefix.
			prefix := strings.TrimPrefix(dir, b.dir)

			runname = addTaskPrefix(prefix, runname)

			// Alter the runname.
			run.SetName(runname)

			// Rewrite dependents to global scope.
			dependsOn := []string{}
			for _, dependent := range run.DependsOn {
				dependsOn = append(dependsOn, addTaskPrefix(prefix, dependent))
			}
			run.DependsOn = dependsOn

			a.RTasks[runname] = run
		}
	}

	return a
}
