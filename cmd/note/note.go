// Package note provides the "note" command and its subcommands.
package note

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/internal/cli/config"
)

// Command returns a cobra.Command named "note" used as a parent to subcommands that manage user notes.
func Command() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:               "note",
		Aliases:           []string{"notes"},
		Short:             "Subcommands for managing notes",
		PersistentPreRunE: cli.CreateClient,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the keeper api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		create(),
		list(),
		delete(),
	)

	return cmd
}
