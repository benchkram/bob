package bobgit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var addTestDataPath = "testdata/add"

func TestGitAdd(t *testing.T) {
	type input struct {
		// environment holds a function creating
		// the testing folder structure.
		environment func(dir string)
	}

	type test struct {
		// name of the test and also used
		// for storing expected output
		name  string
		input input
		// execdir determines the dir in which `bob git status` is executed
		// relative to the repo root.
		// reporoot is used when empty.
		execdir string
		target  string
	}

	tests := []test{
		{
			"add_basic_workspace",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
				},
			},
			"",
			".",
		},
		// validate output when a repo is placed in a subfolder instead of the repo root
		{
			"add_repo_inside_subfolder",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))

					repo := filepath.Join(dir, "subfolder", "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
				},
			},
			"",
			".",
		},
		// validate vorrect display of paths in root repo (../).
		{
			"add_exec_from_subfolder",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))

					repo := filepath.Join(dir, "subfolder", "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
				},
			},
			"subfolder",
			".",
		},
	}

	for _, test := range tests {
		dir, err := ioutil.TempDir("", test.name+"-*")
		assert.Nil(t, err)

		statusBeforeFile := test.name + "_before"
		statusAfterFile := test.name + "_after"

		// Don't cleanup in testdir mode
		if !createTestDirs {
			defer os.RemoveAll(dir)
		}

		if debug || createTestDirs {
			println("Using test dir " + dir)
		}

		test.input.environment(dir)

		if createTestDirs {
			continue
		}

		execdir := filepath.Join(dir, test.execdir)

		statusBefore, err := getStatus(execdir)
		assert.Nil(t, err)

		err = executeAdd(execdir, test.target)
		assert.Nil(t, err)

		statusAfter, err := getStatus(execdir)
		assert.Nil(t, err)

		if update {
			err = os.RemoveAll(filepath.Join(addTestDataPath, statusBeforeFile))
			assert.Nil(t, err)
			err = os.RemoveAll(filepath.Join(addTestDataPath, statusAfterFile))
			assert.Nil(t, err)
			err = os.MkdirAll(addTestDataPath, 0775)
			assert.Nil(t, err)
			err = os.WriteFile(filepath.Join(addTestDataPath, statusBeforeFile), []byte(statusBefore.String()), 0664)
			assert.Nil(t, err)
			err = os.WriteFile(filepath.Join(addTestDataPath, statusAfterFile), []byte(statusAfter.String()), 0664)
			assert.Nil(t, err)
			continue
		}

		expectBefore, err := os.ReadFile(filepath.Join(addTestDataPath, statusBeforeFile))
		assert.Nil(t, err, test.name)

		diff := cmp.Diff(statusBefore.String(), string(expectBefore))
		assert.Equal(t, "", diff, statusBeforeFile)

		expectAfter, err := os.ReadFile(filepath.Join(addTestDataPath, statusAfterFile))
		assert.Nil(t, err, test.name)

		diff = cmp.Diff(statusAfter.String(), string(expectAfter))
		assert.Equal(t, "", diff, statusAfterFile)
	}

	if createTestDirs || update {
		t.FailNow()
	}
}

// executeAdd changes the current working dir before
// executing add command.
func executeAdd(dir string, target string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	defer func() { _ = os.Chdir(wd) }()

	return Add(target)
}
