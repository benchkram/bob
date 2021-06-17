package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Benchkram/bob/bobgit"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/errz"
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

var CmdGitStatus = &cobra.Command{
	Use:   "status",
	Short: "Run git status on all child repos",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runGitStatus()
	},
}

func runGitStatus() {
	s, err := bobgit.Status()
	if err != nil {
		if errors.Is(err, bobutil.ErrCouldNotFindBobDir) {
			fmt.Println("fatal: not a bob repository (or any of the parent directories): .bob")
			os.Exit(1)
		}
		if errors.Is(err, bobgit.ErrCouldNotFindGitDir) {
			fmt.Println("fatal: bobroot is not a git repository")
			os.Exit(1)
		}
		errz.Fatal(err)
	}
	fmt.Println(s.String())
}
