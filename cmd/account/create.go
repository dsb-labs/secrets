package account

import (
	"errors"
	"fmt"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/pkg/secrets"
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

			if password = envvar.String("SECRETS_PASSWORD", ""); password == "" {
				password, err = cli.PromptPassword("Enter password")
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}

				confirm, err := cli.PromptPassword("Confirm password")
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}

				if password != confirm {
					return errors.New("passwords do not match")
				}
			}

			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			account := secrets.Account{
				Email:       args[0],
				DisplayName: args[1],
				Password:    password,
			}

			restoreKey, err := client.CreateAccount(ctx, account)
			if err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}

			fmt.Printf("Created account for %q. Please store the following key for account recovery:\n\n%s\n", account.Email, restoreKey)

			return nil
		},
	}
}
