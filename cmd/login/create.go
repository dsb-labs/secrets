package login

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/davidsbond/passwords/internal/cli"
	"github.com/davidsbond/passwords/pkg/passwords"
)

func create() *cobra.Command {
	var (
		domains []string
	)

	cmd := &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new login",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println("")

			login := passwords.Login{
				Username: args[0],
				Password: string(password),
				Domains:  domains,
			}

			if err = client.CreateLogin(ctx, login); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVar(&domains, "domains", nil, "domains where this login can be used")

	return cmd
}
