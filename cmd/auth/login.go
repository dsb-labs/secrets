package auth

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/davidsbond/keeper/internal/cli"
	"github.com/davidsbond/keeper/internal/cli/config"
)

func login() *cobra.Command {
	return &cobra.Command{
		Use:   "login [email]",
		Short: "Unlock your keeper database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client := cli.ClientFromContext(ctx)

			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(cli.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println("")

			if err = client.Login(ctx, args[0], string(password)); err != nil {
				return fmt.Errorf("failed to login: %w", err)
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

			cfg.Token = client.Token()
			if err = config.Save(configPath, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			return nil
		},
	}
}
