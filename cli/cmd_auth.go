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

const defaultContext = "default"

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Log in to a bob server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		errz.Fatal(err)
	},
}

var AuthContextCreateCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a authentication context",
	Long: `Initialize a authentication context  

Example:
  bob auth init --token=xxx              (default context)
  bob auth init contextName --token=xxx
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := defaultContext
		if len(args) == 1 {
			name = args[0]
		}

		token, err := cmd.Flags().GetString("token")
		if err != nil {
			boblog.Log.UserError(errors.WithMessage(err, "failed to parse token"))
			return
		}

		if token == "" {
			boblog.Log.UserError(fmt.Errorf("token missing"))
			return
		}

		err = runAuthContextCreate(name, token)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
			return
		} else {
			errz.Fatal(err)
		}

		err = runAuthContextSwitch(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
}

var AuthContextUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates an authentication context's token",
	Long: `Updates an authentication context's token  

Example:
  bob auth update --token=xxx              (default context)
  bob auth update contextName --token=xxx
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := defaultContext
		if len(args) == 1 {
			name = args[0]
		}

		token, err := cmd.Flags().GetString("token")
		if err != nil {
			boblog.Log.UserError(errors.WithMessage(err, "failed to parse token"))
			return
		}

		if token == "" {
			boblog.Log.UserError(fmt.Errorf("token missing "))
			return
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
	Use:   "rm",
	Short: "Removes an authentication context",
	Long: `Removes an authentication context  

Example:
  bob auth rm contextName
`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		if name == "" {
			err := cmd.Help()
			errz.Fatal(err)
			return
		}

		err := runAuthContextDelete(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		contextNames, err := getAuthContextNames()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return contextNames, cobra.ShellCompDirectiveDefault
	},
}

var AuthContextSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switches to a different authentication context",
	Long: `Switches to a different authentication context

Example:
  bob auth switch contextName
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		}

		if name == "" {
			err := cmd.Help()
			errz.Fatal(err)
			return
		}

		err := runAuthContextSwitch(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		contextNames, err := getAuthContextNames()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return contextNames, cobra.ShellCompDirectiveDefault
	},
}

func getAuthContextNames() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	authContexts, err := b.AuthContexts()
	errz.Fatal(err)

	if len(authContexts) == 0 {
		return nil, nil
	}

	var names []string
	for _, v := range authContexts {
		names = append(names, v.Name)
	}
	return names, nil
}

var AuthContextListCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists authentication contexts",
	Long:  ``,
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
			curr = " (current)"
		}
		boblog.Log.V(1).Info(fmt.Sprint(c.Name, curr))
	}

	return nil
}
