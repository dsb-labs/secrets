// Package serve provides the "serve" command.
package serve

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/server"
)

// Command returns a cobra.Command named "serve" used to start the password manager server.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "serve [config-file]",
		Short: "Run the server",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			config := server.DefaultConfig()
			if len(args) > 0 {
				config, err = server.LoadConfig(args[0])
				if err != nil {
					return fmt.Errorf("failed to load configuration file: %w", err)
				}
			}

			return server.Run(cmd.Context(), config)
		},
	}
}
