package login

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/passwords/internal/cli"
	"github.com/davidsbond/passwords/internal/cli/config"
)

func list() *cobra.Command {
	var (
		apiURL     string
		configPath string
		domain     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all logins",
		Long:  "Outputs all logins stored for the current user in JSON format.",
		Args:  cobra.NoArgs,
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
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the passwords api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")
	flags.StringVar(&domain, "domain", "", "filter results by domain")

	return cmd
}
