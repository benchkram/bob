package cli

import (
	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
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
	errz.Log(err)

	err = b.RunList()
	errz.Log(err)
}
