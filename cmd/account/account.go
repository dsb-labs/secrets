// Package account provides the "account" command and its subcommands.
package account

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/internal/cli/config"
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
	flags.StringVar(&apiURL, "api-url", envvar.String("KEEPER_API_URL", "http://localhost:8080"), "base url of the keeper api")
	flags.StringVar(&configPath, "config", envvar.String("KEEPER_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		create(),
		info(),
		changePassword(),
		delete(),
		restore(),
	)

	return cmd
}
