package login

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/pkg/secrets"
)

func create() *cobra.Command {
	var (
		domains []string
		name    string
	)

	cmd := &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new login",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			password, err := cli.PromptPassword("Enter password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			login := secrets.Login{
				Username: args[0],
				Password: password,
				Domains:  domains,
				Name:     name,
			}

			id, err := client.CreateLogin(ctx, login)
			if err != nil {
				return err
			}

			fmt.Println(id)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&domains, "domains", "d", nil, "domains where this login can be used")
	flags.StringVarP(&name, "name", "n", "", "login name")

	return cmd
}
