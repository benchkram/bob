package cli

import (
	"fmt"

	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
)

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage remote artifact store authentication contexts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		errz.Fatal(err)
	},
}

var AuthContextCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new authentication context",
	Long:    ``,
	Aliases: []string{"new", "add", "init"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		token, err := cmd.Flags().GetString("token")
		if err != nil {
			boblog.Log.UserError(errors.WithMessage(err, "failed to parse token"))
			return
		}

		if name == "" || token == "" {
			err := cmd.Help()
			errz.Fatal(err)
		}

		err = runAuthContextCreate(name, token)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

var AuthContextUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Updates an authentication context's token",
	Long:    ``,
	Aliases: []string{"set", "set-token"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		token, err := cmd.Flags().GetString("token")
		if err != nil {
			boblog.Log.UserError(errors.WithMessage(err, "failed to parse token"))
			return
		}

		if name == "" || token == "" {
			err := cmd.Help()
			errz.Fatal(err)
		}

		err = runAuthContextUpdate(name, token)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

var AuthContextDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes an authentication context",
	Long:    ``,
	Aliases: []string{"remove", "del", "rm"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		if name == "" {
			err := cmd.Help()
			errz.Fatal(err)
		}

		err := runAuthContextDelete(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

var AuthContextSwitchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switches to a different authentication context",
	Long:    ``,
	Aliases: []string{"select", "set-current"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		if name == "" {
			err := cmd.Help()
			errz.Fatal(err)
		}

		err := runAuthContextSwitch(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

var AuthContextListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists all authentication contexts",
	Long:    ``,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		err := runAuthContextList()
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

func runAuthContextCreate(name, token string) (err error) {
	defer errz.Recover(&err)

	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.CreateAuthContext(name, token)
	errz.Fatal(err)

	boblog.Log.V(1).Info(fmt.Sprintf("Context '%s' created.", name))
	return nil
}

func runAuthContextDelete(name string) (err error) {
	defer errz.Recover(&err)

	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.DeleteAuthContext(name)
	errz.Fatal(err)

	boblog.Log.V(1).Info(fmt.Sprintf("Context '%s' deleted.", name))
	return nil
}

func runAuthContextSwitch(name string) (err error) {
	defer errz.Recover(&err)

	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.SetCurrentAuthContext(name)
	errz.Fatal(err)

	boblog.Log.V(1).Info(fmt.Sprintf("Switched to '%s' context.", name))
	return nil
}

func runAuthContextUpdate(name, token string) (err error) {
	defer errz.Recover(&err)

	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.UpdateAuthContext(name, token)
	errz.Fatal(err)

	boblog.Log.V(1).Info(fmt.Sprintf("Context '%s' updated.", name))
	return nil
}

func runAuthContextList() (err error) {
	defer errz.Recover(&err)

	b, err := bob.Bob()
	errz.Fatal(err)

	ctxs, err := b.AuthContexts()
	errz.Fatal(err)

	if len(ctxs) == 0 {
		boblog.Log.V(1).Info("(empty)")

		return nil
	}

	for _, c := range ctxs {
		var curr string
		if c.Current {
			curr = fmt.Sprintf(" (current)")
		}
		boblog.Log.V(1).Info(fmt.Sprint(c.Name, curr))
	}

	return nil
}
