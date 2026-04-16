// Package account provides the "account" command and its subcommands.
package account

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/internal/cli/config"
)

// Command returns a cobra.Command named "account" used as a parent to subcommands that manage user accounts.
func Command() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:               "account",
		Aliases:           []string{"accounts"},
		Short:             "Subcommands for managing accounts",
		PersistentPreRunE: cli.CreateClient,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&apiURL, "api-url", envvar.String("SECRETS_API_URL", "http://localhost:8080"), "base url of the secrets api")
	flags.StringVar(&configPath, "config", envvar.String("SECRETS_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		create(),
		info(),
		changePassword(),
		delete(),
		restore(),
	)

	return cmd
}
