package account

import (
	"fmt"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/pkg/keeper"
)

func create() *cobra.Command {
	return &cobra.Command{
		Use:   "create [email] [display name]",
		Short: "Create a new account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				password string
				err      error
			)

			if password = envvar.String("KEEPER_PASSWORD", ""); password == "" {
				password, err = cli.PromptPassword()
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
			}

			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			account := keeper.Account{
				Email:       args[0],
				DisplayName: args[1],
				Password:    password,
			}

			if err = client.CreateAccount(ctx, account); err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}

			return nil
		},
	}
}
