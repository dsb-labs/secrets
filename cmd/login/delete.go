package login

import (
	"errors"
	"fmt"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"

	"github.com/davidsbond/passwords/internal/cli/config"
	"github.com/davidsbond/passwords/pkg/passwords"
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

			cfg, err := config.Load(configPath)
			switch {
			case errors.Is(err, config.ErrNotFound):
				return fmt.Errorf("config file not found at %q", configPath)
			case err != nil:
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := passwords.NewClient(apiURL)
			client.SetToken(cfg.Token)

			if err = client.DeleteLogin(ctx, args[0]); err != nil {
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
