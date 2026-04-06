package auth

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/internal/cli/config"
)

func logout() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Lock your secrets database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client := cli.ClientFromContext(ctx)

			if err := client.Logout(ctx); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			cfg, err := config.Load(configPath)
			switch {
			case errors.Is(err, config.ErrNotFound):
				// If a config file has never been created, we'll create a fresh
				// one with the token.
				break
			case err != nil:
				return fmt.Errorf("failed to load config: %w", err)
			}

			cfg.Token = ""
			if err = config.Save(configPath, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			return nil
		},
	}
}
