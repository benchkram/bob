package cli

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify Bobfile",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		//repoURL := args[0]
		runVerify()
	},
}

func runVerify() {
	b, err := bob.Bob()
	errz.Log(err)

	err = b.Verify(context.Background())
	errz.Log(err)

	fmt.Printf("%s\n", aurora.Green("Verified"))
}
