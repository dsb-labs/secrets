// Package login provides the "login" command and its subcommands.
package login

import (
	"github.com/spf13/cobra"
)

// Command returns a cobra.Command named "login" used as a parent to subcommands that manage user logins.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Subcommands for managing logins",
	}

	cmd.AddCommand(
		create(),
		list(),
		delete(),
	)

	return cmd
}
