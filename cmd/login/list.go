package login

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func list() *cobra.Command {
	var (
		domain string
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

			logins, err := client.ListLogins(ctx, domain)
			if err != nil {
				return fmt.Errorf("failed to list logins: %w", err)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			return encoder.Encode(logins)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&domain, "domain", "d", "", "filter results by domain")

	return cmd
}
