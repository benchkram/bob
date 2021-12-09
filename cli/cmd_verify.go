package cli

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify bob.yaml files in a workspace",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		//repoURL := args[0]
		runVerify()
	},
}

func runVerify() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	err = b.Verify(context.Background())
	boblog.Log.Error(err, "Verification failed")

	fmt.Printf("%s\n", aurora.Green("Verified"))
}
