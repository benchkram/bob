//go:build dev
// +build dev

package cli

import (
	"fmt"

	"github.com/benchkram/errz"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getTasksCmd)
}

// getTasksCmd cmd help to profile cli completion
var getTasksCmd = &cobra.Command{
	Use:   "gettasks",
	Short: "gettasks",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := getTasks()
		errz.Log(err)

		for _, t := range tasks {
			fmt.Println(t)
		}
	},
}
