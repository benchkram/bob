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
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

var sshSeverStoragePath = "/tmp/soft"

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
					assert.Nil(t, addAllWithCommit(dir, "Initial Commit"))
					err = cmdutil.RunGitPipe(dir, "push", "-u", "origin", "master")
					assert.Nil(t, err)
					assert.Nil(t, os.WriteFile(filepath.Join(dir, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(dir, "New File Added"))
				},
			},
			"",
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
					err := cmdutil.RunGit(dir, "push", "-u", "origin", "master")
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
					err = cmdutil.RunGit(repo, "push", "-u", "origin", "master")
					assert.Nil(t, err)

					assert.Nil(t, os.WriteFile(filepath.Join(repo, "file2"), []byte("file"), 0664))
					assert.Nil(t, addAllWithCommit(repo, "New File Added"))

				},
			},
			"",
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
					err := cmdutil.RunGit(dir, "push", "-u", "origin", "master")
					assert.Nil(t, err)

					u, _ = url.Parse(sshpath)
					u.Path = path.Join(u.Path, testname+"_repo")

					repo := filepath.Join(dir, "repo")
					assert.Nil(t, os.MkdirAll(repo, 0775))
					err = createGitDirWithRemote(repo, u.String())
					assert.Nil(t, err)
					assert.Nil(t, addAllWithCommit(repo, "Initial Commit"))
					err = cmdutil.RunGit(repo, "push", "-u", "origin", "master")
					assert.Nil(t, err)
				},
			},
			"",
		},
	}

	err := initServer()
	assert.Nil(t, err)

	_, err = cmdutil.RunAddToKnownHost("localhost", cfg.Port)
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

		_, err = getStatus(execdir)
		assert.Nil(t, err)

		err = executePush(execdir)
		assert.Nil(t, err)
	}

	assert.Nil(t, stopServer())
	assert.Nil(t, cleanups())
}

func initServer() error {
	s = server.NewServer(cfg)
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

	err = os.MkdirAll(repopath, 0777)
	if err != nil {
		errz.Fatal(err)
	}

	err = os.MkdirAll(sshpath, 0777)
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

	return Push()
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
