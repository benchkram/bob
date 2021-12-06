package bobgit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Benchkram/bob/bobgit/status"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var update = false
var debug = false

// createTestDirs structure without runnign tests.
// Usful to debug the created repo structure.
var createTestDirs = false

func TestStatus(t *testing.T) {

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
	}

	tests := []test{
		{
			"status_untracked",
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
		},
		// validate output when a repo is placed in a subfolder instead of the repo root
		{
			"status_subfolder",
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
		},
		// validate vorrect display of paths in root repo (../).
		{
			"status_exec_from_subfolder",
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
		},
		{
			"status_added",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
				},
			},
			"",
		},
		{
			"status_comitted",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "initialcommit"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("changedfilecontent"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(repo, "commit", "-m", "initialcommit"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("changedfilecontent"), 0664))
				},
			},
			"",
		},
		{
			"status_conflict",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "initialcommit"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("changedfilecontent"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content"))

					assertMergeConflict(t, dir)
				},
			},
			"",
		},
		{
			"status_conflict_multirepo",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "initialcommit"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("changedfilecontent"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content"))

					assert.Nil(t, cmdutil.RunGit(dir, "checkout", "-b", "target_branch"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file content changed in target branch"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content from target branch"))

					assert.Nil(t, cmdutil.RunGit(dir, "checkout", "master"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file content changed in main branch"), 0664))
					assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content from target branch"))
					assert.NotNil(t, cmdutil.RunGit(dir, "merge", "target_branch"))

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(repo, "commit", "-m", "initialcommit"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("changedfilecontent"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(repo, "commit", "-m", "Updated content"))

					assert.Nil(t, cmdutil.RunGit(repo, "checkout", "-b", "target_branch"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file content changed in target branch"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(repo, "commit", "-m", "Updated content from target branch"))

					assert.Nil(t, cmdutil.RunGit(repo, "checkout", "master"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file content changed in main branch"), 0664))
					assert.Nil(t, cmdutil.RunGit(repo, "add", "--all"))
					assert.Nil(t, cmdutil.RunGit(repo, "commit", "-m", "Updated content from target branch"))
					assert.NotNil(t, cmdutil.RunGit(repo, "merge", "target_branch"))
				},
			},
			"",
		},
	}

	for _, test := range tests {
		dir, err := ioutil.TempDir("", test.name+"-*")
		assert.Nil(t, err)

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

		status, err := getStatus(filepath.Join(dir, test.execdir))
		assert.Nil(t, err)

		if update {
			err = os.RemoveAll(filepath.Join("testdata", test.name))
			assert.Nil(t, err)
			err = os.MkdirAll("testdata", 0775)
			assert.Nil(t, err)
			err = os.WriteFile(filepath.Join("testdata", test.name), []byte(status.String()), 0664)
			assert.Nil(t, err)
			continue
		}

		expect, err := os.ReadFile(filepath.Join("testdata", test.name))
		assert.Nil(t, err, test.name)

		diff := cmp.Diff(status.String(), string(expect))
		assert.Equal(t, "", diff, test.name)
	}

	if createTestDirs || update {
		t.FailNow()
	}
}

// getStatus changes the current working dir before
// executing retireving the status.
func getStatus(dir string) (s *status.S, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Chdir(wd) }()

	return Status()
}

func assertMergeConflict(t *testing.T, dir string) {
	assert.Nil(t, cmdutil.RunGit(dir, "checkout", "-b", "target_branch"))
	assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file content changed in target branch"), 0664))
	assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
	assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content from target branch"))

	assert.Nil(t, cmdutil.RunGit(dir, "checkout", "master"))
	assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file content changed in main branch"), 0664))
	assert.Nil(t, cmdutil.RunGit(dir, "add", "--all"))
	assert.Nil(t, cmdutil.RunGit(dir, "commit", "-m", "Updated content from target branch"))
	assert.NotNil(t, cmdutil.RunGit(dir, "merge", "target_branch"))
}
