package bob

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// syncProjectName project names for all bobfiles and build tasks
func syncProjectName(a *bobfile.Bobfile, bobs []*bobfile.Bobfile) (*bobfile.Bobfile, []*bobfile.Bobfile) {
	toSync := append([]*bobfile.Bobfile{a}, bobs...)

	for _, bobfile := range toSync {
		bobfile.Project = a.Project

		for taskname, task := range bobfile.BTasks {
			// Name of the umbrella-bobfile.
			task.SetProject(a.Project)

			// Overwrite value in build map
			bobfile.BTasks[taskname] = task
		}
	}

	return a, bobs
}

func (b *B) addBuildTasksToAggregate(a *bobfile.Bobfile, bobs []*bobfile.Bobfile, decorations map[string][]string) (*bobfile.Bobfile, error) {
	allTasks := make(map[string]bool)

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

			allTasks[taskname] = true
			// Alter the taskname.
			task.SetName(taskname)

			// Rewrite dependent tasks to global scope.
			var dependsOn []string
			if dependsFromDecoration, ok := decorations[taskname]; ok {
				dependsOn = append(dependsOn, dependsFromDecoration...)
			}
			for _, dependentTask := range task.DependsOn {
				dependsOn = append(dependsOn, addTaskPrefix(prefix, dependentTask))
			}
			task.DependsOn = dependsOn

			a.BTasks[taskname] = task
		}
	}

	// validate if child task exists for decoration
	for k := range decorations {
		if _, ok := allTasks[k]; !ok {
			return a, usererror.Wrap(fmt.Errorf("you are modifing an imported task `%s` that does not exist", k))
		}
	}

	return a, nil
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

// readImports recursively
//
// readModePlain allows to read bobfiles without
// doing sanitization.
//
// If prefix is given it's appended to the search path to assure
// correctness of the search path in case of recursive calls.
func readImports(
	a *bobfile.Bobfile,
	readModePlain bool,
	prefix ...string,
) (imports []*bobfile.Bobfile, err error) {
	errz.Recover(&err)

	var p string
	if len(prefix) > 0 {
		p = prefix[0]
	}

	imports = []*bobfile.Bobfile{}
	for _, importPath := range a.Imports {
		// read bobfile
		var boblet *bobfile.Bobfile
		var err error
		if readModePlain {
			boblet, err = bobfile.BobfileReadPlain(filepath.Join(p, importPath))
		} else {
			boblet, err = bobfile.BobfileRead(filepath.Join(p, importPath))
		}
		if err != nil {
			if errors.Is(err, bobfile.ErrBobfileNotFound) {
				return nil, usererror.Wrapm(err, fmt.Sprintf("import of %s from %s/bob.yaml failed", importPath, a.Dir()))
			}
			errz.Fatal(err)
		}
		imports = append(imports, boblet)

		// read imports recursively
		childImports, err := readImports(boblet, readModePlain, boblet.Dir())
		errz.Fatal(err)
		imports = append(imports, childImports...)
	}

	return imports, nil
}
