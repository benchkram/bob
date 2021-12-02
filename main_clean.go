package main

import (
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean buildinfo",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runClean()
	},
}

func runClean() {
	b, err := bob.Bob()
	errz.Log(err)

	err = b.Clean()
	errz.Log(err)

	fmt.Println("build info store cleaned")
	fmt.Println("local artifact store cleaned")
}
