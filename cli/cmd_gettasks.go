//go:build dev
// +build dev

package cli

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"

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
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}

		for _, t := range tasks {
			fmt.Println(t)
		}
	},
}
