// Package auth provides the "auth" command and its subcommands.
package auth

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/internal/cli/config"
)

// Command returns a cobra.Command named "auth" used as a parent to subcommands that manage user authentication.
func Command() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:               "auth",
		Short:             "Subcommands for authentication",
		PersistentPreRunE: cli.CreateClient,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&apiURL, "api-url", envvar.String("KEEPER_API_URL", "http://localhost:8080"), "base url of the secrets api")
	flags.StringVar(&configPath, "config", envvar.String("KEEPER_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		login(),
		logout(),
	)

	return cmd
}
