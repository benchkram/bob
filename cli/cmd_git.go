package cli

import (
	"errors"
	"fmt"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bobgit"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/bobutil"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

var CmdGit = &cobra.Command{
	Use:   "git",
	Short: "Run git cmds on all child repos",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		errz.Fatal(err)
	},
}

var CmdGitAdd = &cobra.Command{
	Use:   "add",
	Short: "Run git add on all child repos",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runGitAdd(args...)
	},
}

var CmdGitCommit = &cobra.Command{
	Use:   "commit",
	Short: "Run git commit on all child repos using the given message",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		message, _ := cmd.Flags().GetString("message")
		runGitCommit(message)
	},
}

var CmdGitStatus = &cobra.Command{
	Use:   "status",
	Short: "Run git status on all child repos",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runGitStatus()
	},
}

func runGitAdd(targets ...string) {
	err := bobgit.Add(targets...)
	if err != nil {
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
			exit(1)
		} else {
			errz.Fatal(err)
		}
	}
}

func runGitCommit(m string) {
	s, err := bobgit.Commit(m)
	if err != nil {
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
			exit(1)
		} else if errors.Is(err, bobgit.ErrEmptyCommitMessage) {
			fmt.Printf("%s\n\n  %s\n\n", "bob git requires a commit message", aurora.Bold("bob git commit -m \"msg\""))
			exit(1)
		} else {
			errz.Fatal(err)
		}
	}

	if s != "" {
		fmt.Println(s)
	}
}

func runGitStatus() {
	s, err := bobgit.Status()
	if err != nil {
		if errors.Is(err, bobutil.ErrCouldNotFindBobWorkspace) {
			fmt.Println("fatal: not a bob repository (or any of the parent directories): .bob")
			exit(1)
		}
		if errors.Is(err, bobgit.ErrCouldNotFindGitDir) {
			fmt.Println("fatal: bob workspace is not a git repository")
			exit(1)
		}
		errz.Fatal(err)
	}
	fmt.Println(s.String())
}
