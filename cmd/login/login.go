// Package login provides the "login" command and its subcommands.
package login

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/internal/cli/config"
)

// Command returns a cobra.Command named "login" used as a parent to subcommands that manage user logins.
func Command() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:               "login",
		Aliases:           []string{"logins"},
		Short:             "Subcommands for managing logins",
		PersistentPreRunE: cli.CreateClient,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the keeper api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		create(),
		list(),
		delete(),
		get(),
	)

	return cmd
}
