package bobgit

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

var addTestDataPath = "testdata/add"

// disable running tests for add,
// helpful while writing other tests
var runAddTests = true

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

					createGitRepo(t, dir, "repo")
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
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/"), 0664))

					createGitRepo(t, dir, "subfolder", "repo")
				},
			},
			"",
			".",
		},
		// validate the addition of files starting from the subdirectory and
		// all the repository under that subdirectory.
		{
			"add_exec_from_subfolder",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					createGitRepo(t, subfolder, "repo")
					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder",
			".",
		},
		// should add all the files from the repo and second level and ignore other files
		{
			"add_exec_from_repo_inside_subfolder",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder/repo",
			".",
		},
		// should add only files from repo2
		{
			"add_exec_all_from_second_repo_inside_subfolder",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder/repo2",
			".",
		},
		// execute add all to the root from the deepest respository
		{
			"add_all_exec_to_root_from_deepest_repo",
			input{
				func(dir string) {
					err := cmdutil.RunGit(dir, "init")
					assert.Nil(t, err)

					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder/repo/second-level",
			"../../../.",
		},
		// execute add all to the subfolder from the deepest respository
		// skips the files from the root directory
		{
			"add_all_exec_to_subfolder_from_deepest_repo",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder/repo/second-level",
			"../../.",
		},
		// execute add all to the subfolder from the deepest respository
		// skips the files from the root directory
		{
			"add_all_exec_to_subfolder_from_deepest_repo",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					createGitRepo(t, subfolder, "repo2")
				},
			},
			"subfolder/repo/second-level",
			"../../.",
		},
		// add all repository starting from one repository
		// target specific repository directory instead of `.`
		{
			"add_repository_dir_from_root",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					createGitRepo(t, subfolder, "repo2")
				},
			},
			"",
			"subfolder/repo/.",
		},
		// select all files ending with .txt from repo2
		{
			"add_all_files_ending_txt_from_single_repo",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					repo2Path := filepath.Join(subfolder, "repo2")
					createGitRepo(t, repo2Path)
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file.txt"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file2.txt"), []byte("file"), 0664))
				},
			},
			"",
			`subfolder/repo2/*.txt`,
		},
		// add a single file from the list of bob git status
		{
			"add_single_file_with_backward_path",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					repo2Path := filepath.Join(subfolder, "repo2")
					createGitRepo(t, repo2Path)
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file.txt"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file2.txt"), []byte("file"), 0664))
				},
			},
			"subfolder/repo",
			"../../subfolder/repo2/file",
		},
		// add multiple pathspecs
		{
			"add_multiple_pathspecs_with_regex",
			input{
				func(dir string) {
					createGitRepo(t, dir)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("subfolder/repo/\nsubfolder/repo2/"), 0664))

					subfolder := filepath.Join(dir, "subfolder")
					assert.Nil(t, os.MkdirAll(subfolder, 0775))
					assert.Nil(t, os.WriteFile(filepath.Join(subfolder, "folder-file"), []byte("file"), 0664))

					repoPath := filepath.Join(subfolder, "repo")
					createGitRepo(t, repoPath)
					assert.Nil(t, os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("second-level/"), 0664))
					createGitRepo(t, repoPath, "second-level")

					repo2Path := filepath.Join(subfolder, "repo2")
					createGitRepo(t, repo2Path)
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file.txt"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(repo2Path, "file2.txt"), []byte("file"), 0664))
				},
			},
			"",
			"subfolder/repo/second-level/ subfolder/repo2/*.txt",
		},
	}

	if runAddTests {

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

			targets := strings.Split(test.target, " ")
			err = executeAdd(execdir, targets...)
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
	}

	if createTestDirs || update {
		t.FailNow()
	}
}

// executeAdd changes the current working dir before
// executing add command.
func executeAdd(dir string, target ...string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	defer func() { _ = os.Chdir(wd) }()

	return Add(target...)
}

func createGitRepo(t *testing.T, dirs ...string) {
	repo := filepath.Join(dirs...)
	assert.Nil(t, os.MkdirAll(repo, 0775))
	assert.Nil(t, cmdutil.RunGit(repo, "init"))
	assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
}
