package cli

import (
	"fmt"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/spf13/cobra"
)

var runListCmd = &cobra.Command{
	Use:   "ls",
	Short: "ls",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runRunList()
	},
}

func runRunList() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	tasks, err := b.GetRunTasks()
	boblog.Log.Error(err, "Unable to list tasks")

	for _, t := range tasks {
		fmt.Println(t)
	}
}
