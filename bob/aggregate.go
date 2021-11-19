package bob

import (
	"path/filepath"
	"strings"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/pkg/filepathutil"
)

// find bobfiles recursively.
func (b *B) find() (bobfiles []string, err error) {
	defer errz.Recover(&err)

	list, err := filepathutil.ListRecursive(b.dir)
	errz.Fatal(err)

	for _, file := range list {
		if bobfile.IsBobfile(file) {
			bobfiles = append(bobfiles, file)
		}
	}

	return bobfiles, nil
}

// Aggregate determine and read Bobfiles recursively into memory
// and returns a single Bobfile containing all tasks & runs.
func (b *B) Aggregate() (aggregate *bobfile.Bobfile, err error) {
	defer errz.Recover(&err)

	bobfiles, err := b.find()
	errz.Fatal(err)

	// Read & Find Bobfiles
	bobs := []*bobfile.Bobfile{}
	for _, bf := range bobfiles {
		boblet, err := bobfile.BobfileRead(filepath.Dir(bf))
		errz.Fatal(err)

		if boblet.Dir() == b.dir {
			aggregate = boblet
		}

		for variable, value := range boblet.Variables {
			for key, task := range boblet.Tasks {
				// TODO: Create and use envvar sanitizer
				task.AddEnvironment(strings.ToUpper(variable), value)
				boblet.Tasks[key] = task
			}
		}

		bobs = append(bobs, boblet)
	}

	if aggregate == nil {
		return nil, ErrCouldNotFindTopLevelBobfile
	}

	// Merge tasks into one Bobfile
	for _, bobfile := range bobs {
		// Skip the aggregate
		if bobfile.Dir() == aggregate.Dir() {
			continue
		}

		for taskname, task := range bobfile.Tasks {
			dir := bobfile.Dir()

			// Use a relative path as task prefix.
			prefix := strings.TrimPrefix(dir, b.dir)
			taskname := addTaskPrefix(prefix, taskname)

			// fmt.Printf("aggreagted [dir:%s, bdir:%s prefix:%s] taskname %s\n", prefix, dir, b.dir, taskname)

			// Alter the taskname.
			task.SetName(taskname)

			// Rewrite dependent tasks to global scope.
			dependsOn := []string{}
			for _, dependentTask := range task.DependsOn {
				dependsOn = append(dependsOn, addTaskPrefix(prefix, dependentTask))
			}
			task.DependsOn = dependsOn

			aggregate.Tasks[taskname] = task
		}
	}

	// Merge runs into one Bobfile
	for _, bobfile := range bobs {
		// Skip the aggregate
		if bobfile.Dir() == aggregate.Dir() {
			continue
		}

		for runname, run := range bobfile.Runs {
			dir := bobfile.Dir()

			// Use a relative path as task prefix.
			prefix := strings.TrimPrefix(dir, b.dir)
			name := addTaskPrefix(prefix, runname)

			// Alter the runname.
			run.SetName(name)

			// Rewrite dependents to global scope.
			dependsOn := []string{}
			for _, dependent := range run.DependsOn {
				dependsOn = append(dependsOn, addTaskPrefix(prefix, dependent))
			}
			run.DependsOn = dependsOn

			aggregate.Runs[name] = run
		}
	}

	// Gather environment from dependent tasks.
	//
	// Each export is translated into environment variables named:
	//   `second-level/openapi => SECOND_LEVEL_OPENAPI`
	// hyphens`-` are translated to underscores`_`.
	//
	// The file is prefixed with all paths to make it relative to dir of the the top Bobfile:
	//   `openapi.yaml => sencond-level/openapi.yaml`
	for i, task := range aggregate.Tasks {

		for _, dependentTaskName := range task.DependsOn {

			dependentTask, ok := aggregate.Tasks[dependentTaskName]
			if !ok {
				return nil, ErrTaskDoesNotExist
			}

			for exportname, export := range dependentTask.Exports {
				// fmt.Printf("Task %s exports %s\n", dependentTaskName, export)

				envvar := taskNameToEnvironment(dependentTaskName, exportname)

				value := filepath.Join(dependentTask.Dir(), string(export))

				// Make the path relative to the aggregates dir.
				dir := aggregate.Dir()
				if !strings.HasSuffix(dir, "/") {
					dir = dir + "/"
				}
				value = strings.TrimPrefix(value, dir)

				// println(envvar, value)

				task.AddEnvironment(envvar, value)

				aggregate.Tasks[i] = task
			}
		}
	}

	return aggregate, aggregate.Verify()
}

func addTaskPrefix(prefix, taskname string) string {
	taskname = filepath.Join(prefix, taskname)
	taskname = strings.TrimPrefix(taskname, "/")
	return taskname
}

// taskNameToEnvironment
//
// Each taskname is translated into environment variables like:
//   `second-level/openapi_exportname => SECOND_LEVEL_OPENAPI_EXPORTNAME`
// Hyphens`-` are translated to underscores`_`.
func taskNameToEnvironment(taskname string, exportname string) string {

	splits := strings.Split(taskname, "/")
	splits = append(splits, exportname)

	envvar := strings.Join(splits, "_")
	envvar = strings.ReplaceAll(envvar, "-", "_")
	envvar = strings.ReplaceAll(envvar, ".", "_")
	envvar = strings.ToUpper(envvar)

	return envvar
}
