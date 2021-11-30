package main

import (
	"github.com/spf13/cobra"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
)

var cmdWorkspaceNew = &cobra.Command{
	Use:   "new",
	Short: "Create a new bob workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runWorkspaceNew()
	},
}

func runWorkspaceNew() {
	bob, err := bob.Bob()
	errz.Fatal(err)

	err = bob.Init()
	errz.Fatal(err)
}
