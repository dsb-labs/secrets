// Package note provides the "note" command and its subcommands.
package note

import (
	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/internal/cli/config"
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
	flags.StringVar(&apiURL, "api-url", envvar.String("KEEPER_API_URL", "http://localhost:8080"), "base url of the secrets api")
	flags.StringVar(&configPath, "config", envvar.String("KEEPER_CONFIG", config.DefaultConfigPath()), "path to config file")

	cmd.AddCommand(
		create(),
		list(),
		delete(),
		get(),
	)

	return cmd
}
