package login

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/pkg/keeper"
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

			password, err := cli.PromptPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			login := keeper.Login{
				Username: args[0],
				Password: password,
				Domains:  domains,
			}

			if err = client.CreateLogin(ctx, login); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&domains, "domains", "d", nil, "domains where this login can be used")

	return cmd
}
