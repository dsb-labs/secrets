package login

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/pkg/keeper"
)

func list() *cobra.Command {
	var (
		domain string
		name   string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all logins",
		Long:    "Outputs all logins stored for the current user in JSON format.",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			logins, err := client.ListLogins(ctx, keeper.LoginListOptions{
				Domain: domain,
				Name:   name,
			})
			if err != nil {
				return fmt.Errorf("failed to list logins: %w", err)
			}

			return cli.Write(logins)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&domain, "domain", "d", "", "filter results by domain")
	flags.StringVarP(&name, "name", "n", "", "filter results by name")

	return cmd
}
