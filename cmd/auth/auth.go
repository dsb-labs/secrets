// Package auth provides the "auth" command and its subcommands.
package auth

import (
	"github.com/spf13/cobra"
)

// Command returns a cobra.Command named "auth" used as a parent to subcommands that manage user authentication.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Subcommands for authentication",
	}

	cmd.AddCommand(
		login(),
	)

	return cmd
}
