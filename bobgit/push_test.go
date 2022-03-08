package bobgit

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/errz"
	"github.com/charmbracelet/soft-serve/config"
	"github.com/charmbracelet/soft-serve/server"
	"github.com/gliderlabs/ssh"
	"github.com/google/go-cmp/cmp"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

var sshSeverStoragePath = "/tmp/soft"
var pushTestDataPath = "testdata/push"

// disable running tests for push,
// helpful while writing other tests
var runPushTests = true

var cfg *config.Config
var s *server.Server

func TestPush(t *testing.T) {
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

		// test returned errors
		expectederr error
	}

	cfg = createCustomConfig(sshSeverStoragePath)
	var sshpath string = fmt.Sprintf("ssh://localhost:%s/", fmt.Sprint(cfg.Port))

	tests := []test{
		{
			"single_repo_push",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					err := createGitDirWithRemote(dir, u.String())
					assert.Nil(t, err)

					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "'Initial Commit'"))
					err = cmdutil.GitPushFirstTime(dir, "origin", "master", true)
					if err != nil {
						fmt.Println(err)
					}
					assert.Nil(t, err)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))
				},
			},
			"",
			nil,
		},
		{
			"multi_repo_push",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					assert.Nil(t, createGitDirWithRemote(dir, u.String()))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					err := cmdutil.GitPushFirstTime(dir, "origin", "master", true)
					assert.Nil(t, err)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))

					u, _ = url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname+"_repo")

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					err = createGitDirWithRemote(repo, u.String())
					assert.Nil(t, err)
					assert.Nil(t, addAllWithCommit(repo, "Initial Commit"))
					err = cmdutil.GitPushFirstTime(repo, "origin", "master", true)
					assert.Nil(t, err)

					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(repo, "New File Added"))

				},
			},
			"",
			nil,
		},
		{
			"multi_repo_nothing_to_push",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					assert.Nil(t, createGitDirWithRemote(dir, u.String()))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					err := cmdutil.GitPushFirstTime(dir, "origin", "master", true)
					assert.Nil(t, err)

					u, _ = url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname+"_repo")

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					err = createGitDirWithRemote(repo, u.String())
					assert.Nil(t, err)
					assert.Nil(t, addAllWithCommit(repo, "Initial Commit"))
					err = cmdutil.GitPushFirstTime(repo, "origin", "master", true)
					assert.Nil(t, err)
				},
			},
			"",
			ErrUptodateAllRepo,
		},
		{
			"multi_repo_with_one_repo_push",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					assert.Nil(t, createGitDirWithRemote(dir, u.String()))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					err := cmdutil.GitPushFirstTime(dir, "origin", "master", true)
					assert.Nil(t, err)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))

					u, _ = url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname+"_repo")

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					err = createGitDirWithRemote(repo, u.String())
					assert.Nil(t, err)
					assert.Nil(t, addAllWithCommit(repo, "Initial Commit"))
					err = cmdutil.GitPushFirstTime(repo, "origin", "master", true)
					assert.Nil(t, err)
				},
			},
			"repo",
			nil,
		},
		{
			"single_repo_with_no_configured_remote",
			input{
				func(dir string) {
					assert.Nil(t, os.MkdirAll(dir, 0775))
					assert.Nil(t, cmdutil.RunGit(dir, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
				},
			},
			"",
			ErrInsufficientConfig,
		},
		{
			"multi_repo_with_no_configured_remote",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					assert.Nil(t, createGitDirWithRemote(dir, u.String()))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					err := cmdutil.GitPushFirstTime(dir, "origin", "master", true)
					assert.Nil(t, err)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					assert.Nil(t, cmdutil.RunGit(repo, "init"))
					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(repo, "Initial Commit"))
				},
			},
			"",
			ErrInsufficientConfig,
		},
		// normal bob push should fail, as no branch in remote exists in
		// configured remote
		{
			"single_repo_for_first_time_push",
			input{
				func(dir string) {

					testname := filepath.Base(dir)

					u, _ := url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname)

					assert.Nil(t, createGitDirWithRemote(dir, u.String()))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("repo/"), 0664))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, ".bob.workspace"), []byte(""), 0664))
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))
				},
			},
			"",
			ErrInsufficientConfig,
		},
	}

	if runPushTests {
		err := initServer()
		assert.Nil(t, err)

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

			err = executePush(execdir)
			if test.expectederr != nil {
				if errors.Is(err, test.expectederr) {
					continue
				}
				assert.Fail(t, fmt.Sprintf("expected error [%s] got [%s]", test.expectederr.Error(), err.Error()))
			}

			assert.Nil(t, err)

			statusAfter, err := getStatus(execdir)
			assert.Nil(t, err)

			if update {
				// tests expecting a error don't need to compare their before and after outputs
				if test.expectederr != nil {
					continue
				}

				err = os.RemoveAll(filepath.Join(pushTestDataPath, test.name))
				assert.Nil(t, err)
				err = os.MkdirAll(pushTestDataPath, 0775)
				assert.Nil(t, err)
				err = os.WriteFile(filepath.Join(pushTestDataPath, test.name), []byte(statusAfter.String()), 0664)
				assert.Nil(t, err)
				continue
			}

			expectAfter, err := os.ReadFile(filepath.Join(pushTestDataPath, test.name))
			assert.Nil(t, err, test.name)

			diff := cmp.Diff(statusAfter.String(), string(expectAfter))
			assert.Equal(t, "", diff, test.name)
		}

		assert.Nil(t, stopServer())
		assert.Nil(t, cleanups())
	}

	if createTestDirs || update {
		t.FailNow()
	}
}

func initServer() error {
	s = server.NewServer(cfg)
	fmt.Println("Server is running on " + fmt.Sprint(s.Config.Port))
	go func() {
		err := s.Start()
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			errz.Fatal(err)
		}
	}()

	return nil
}

func stopServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	err := s.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func cleanups() error {
	_, err := cmdutil.RemoveFromKnownHost("localhost", cfg.Port)
	if err != nil {
		return err
	}

	err = os.RemoveAll(sshSeverStoragePath)
	if err != nil {
		return err
	}

	return nil
}

func createCustomConfig(temppath string) *config.Config {
	repopath := filepath.Join(temppath, ".repos")
	sshpath := filepath.Join(temppath, ".ssh")

	port, err := freeport.GetFreePort()
	if err != nil {
		errz.Fatal(err)
	}

	err = os.MkdirAll(repopath, 0775)
	if err != nil {
		errz.Fatal(err)
	}

	err = os.MkdirAll(sshpath, 0775)
	if err != nil {
		errz.Fatal(err)
	}

	keyfile := filepath.Join(sshpath, "soft_serve_server_ed25519")

	cfg := config.DefaultConfig()
	cfg.Port = port
	cfg.RepoPath = repopath
	cfg.KeyPath = keyfile

	return cfg
}

// executePush changes the current working dir before
// executing push command.
func executePush(dir string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	defer func() { _ = os.Chdir(wd) }()

	return Push(EnableTesting())
}

func createGitDirWithRemote(dir string, remotepath string) error {
	if err := os.MkdirAll(dir, 0775); err != nil {
		return err
	}

	if err := cmdutil.RunGit(dir, "init"); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("file"), 0664); err != nil {
		return err
	}

	if err := cmdutil.RunGit(dir, "remote", "add", "origin", remotepath); err != nil {
		return err
	}

	return nil
}

func addAllWithCommit(dir string, message string) error {
	if err := cmdutil.RunGit(dir, "add", "--all"); err != nil {
		return err
	}

	if err := cmdutil.RunGit(dir, "commit", "-m", message); err != nil {
		return err
	}

	return nil
}
