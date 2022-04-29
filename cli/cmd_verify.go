package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
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
		stopProfiling()
		if exitCode == 0 {
			fmt.Printf("\n%s\n", aurora.Green("Verified"))
			exit(0)
		} else {
			fmt.Printf("\n%s\n", aurora.Red("Verification failed"))
			exit(exitCode)
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
