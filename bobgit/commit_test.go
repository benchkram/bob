package bobgit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var commitTestDataPath = "testdata/commit"

// disable running tests for commit,
// helpful while writing other tests
var runCommitTests = true

func TestCommit(t *testing.T) {

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
		// send commit message to test
		message string
		// output string from Commit function
		output string

		expectedErr error
	}

	tests := []test{
		// nothing should be commited, prints a additional user message with untracked repository list
		{
			"commit_untracked",
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
			"test commmits with untracked files",
			UntrackedRepoMessage([]string{".", "repo"}),
			nil,
		},
		// all files should be commited except files in `repo` sub-directory
		{
			"commit_tracked_bobroot",
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

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)
				},
			},
			"",
			"test commmits with tracked files in bobroot",
			"",
			nil,
		},
		// all files should be commited even files in `repo` sub-directory
		{
			"commit_tracked_all",
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
					err = cmdutil.RunGit(repo, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)
				},
			},
			"",
			"test commmits with tracked files",
			"",
			nil,
		},
		// only files in `repo` sub-directory should be commited
		// files in bobroot should be unmodified
		{
			"commit_repo_subdirectory_files",
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
					err = cmdutil.RunGit(repo, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)
					err = cmdutil.RunGit(dir, "commit", "-m", "bobroot changes commited")
					assert.Nil(t, err)
				},
			},
			"",
			"test commmits with tracked files in repo subdir",
			"",
			nil,
		},
		// execute git commit from repo sub repository
		// should work as same as executing from bobroot
		{
			"exec_commit_from_subdrepository",
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
					err = cmdutil.RunGit(repo, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)
				},
			},
			"repo",
			"test commmits from sub-repository",
			"",
			nil,
		},
		// exec commit from subdirectory
		// should work as same as executing from bobroot
		{
			"exec_commit_from_subdirectory",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))

					repo := filepath.Join(dir, "subdirectory", "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
					err = cmdutil.RunGit(repo, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)
				},
			},
			"subdirectory",
			"test commmits from subdirectory",
			"",
			nil,
		},
		// exec commit without a message
		// execution should return a ErrEmptyCommitMessage error
		{
			"exec_commit_without_message",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
				},
			},
			"",
			"",
			"",
			ErrEmptyCommitMessage,
		},
		// exec commit without a nothing to update
		// execution should return a User message about nothing to update
		{
			"exec_commit_after_commiting_updates",
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

					err = cmdutil.RunGit(repo, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(repo, "commit", "-m", "All changes commited")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "commit", "-m", "All changes commited")
					assert.Nil(t, err)
				},
			},
			"",
			"exec commit after commiting one time",
			CleanWorkingDirMessage,
			nil,
		},

		// exec commit without a tracking on subrepository
		// execution should return a User message about untracked repository name
		{
			"exec_commit_without_tracking_one_repo",
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

					err = cmdutil.RunGit(dir, "add", ".")
					assert.Nil(t, err)

					err = cmdutil.RunGit(dir, "commit", "-m", "All changes commited")
					assert.Nil(t, err)
				},
			},
			"",
			"exec commit after commiting one time",
			UntrackedRepoMessage([]string{"repo"}),
			nil,
		},
	}

	if runCommitTests {

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

			execdir := filepath.Join(dir, test.execdir)
			statusBeforeFile := test.name + "_before"
			statusAfterFile := test.name + "_after"

			statusBefore, err := getStatus(execdir)
			assert.Nil(t, err)

			s, err := executeCommit(execdir, test.message)
			if test.expectedErr != nil {
				if errors.Is(err, test.expectedErr) {
					continue
				}
				assert.Fail(t, fmt.Sprintf("expected error [%s] got [%s]", test.expectedErr.Error(), err.Error()))
			}

			// ignore the error caused by test.message nill
			if err != nil && !errors.Is(err, ErrEmptyCommitMessage) {
				assert.Nil(t, err)
			}

			assert.Equal(t, s, test.output, test.name)

			statusAfter, err := getStatus(execdir)
			assert.Nil(t, err)

			if update {
				// tests expecting a error don't need to compare their before and after putputs
				if test.expectedErr != nil {
					continue
				}

				err = os.RemoveAll(filepath.Join(commitTestDataPath, statusBeforeFile))
				assert.Nil(t, err)
				err = os.RemoveAll(filepath.Join(commitTestDataPath, statusAfterFile))
				assert.Nil(t, err)
				err = os.MkdirAll(commitTestDataPath, 0775)
				assert.Nil(t, err)
				err = os.WriteFile(filepath.Join(commitTestDataPath, statusBeforeFile), []byte(statusBefore.String()), 0664)
				assert.Nil(t, err)
				err = os.WriteFile(filepath.Join(commitTestDataPath, statusAfterFile), []byte(statusAfter.String()), 0664)
				assert.Nil(t, err)
				continue
			}

			expectBefore, err := os.ReadFile(filepath.Join(commitTestDataPath, statusBeforeFile))
			assert.Nil(t, err, test.name)

			diff := cmp.Diff(statusBefore.String(), string(expectBefore))
			assert.Equal(t, "", diff, statusBeforeFile)

			expectAfter, err := os.ReadFile(filepath.Join(commitTestDataPath, statusAfterFile))
			assert.Nil(t, err, test.name)

			diff = cmp.Diff(statusAfter.String(), string(expectAfter))
			assert.Equal(t, "", diff, statusAfterFile)
		}
	}

	if createTestDirs || update {
		t.FailNow()
	}
}

// executeCommit changes the current working dir before
// executing commit command.
func executeCommit(dir string, message string) (_ string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = os.Chdir(dir)
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Chdir(wd) }()

	return Commit(message)
}
