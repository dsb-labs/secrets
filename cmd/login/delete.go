package login

import (
	"fmt"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/passwords/internal/cli"
	"github.com/davidsbond/passwords/internal/cli/config"
)

func delete() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a login",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			if err := client.DeleteLogin(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete login: %w", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the passwords api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")

	return cmd
}
