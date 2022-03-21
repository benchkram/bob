package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify bob.yaml files in a workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runVerify()
	},
}

func runVerify() {
	exitCode := 0
	defer func() {
		if exitCode == 0 {
			fmt.Printf("\n%s\n", aurora.Green("Verified"))
			os.Exit(0)
		} else {
			fmt.Printf("\n%s\n", aurora.Red("Verification failed"))
			os.Exit(exitCode)
		}
	}()

	b, err := bob.Bob()
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Log(err)
		}
	}

	err = b.Verify(context.Background())
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Log(err)
		}
	}
}
