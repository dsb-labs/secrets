package auth

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/davidsbond/keeper/internal/cli/config"
	"github.com/davidsbond/keeper/pkg/keeper"
)

func login() *cobra.Command {
	var (
		apiURL     string
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "login [email]",
		Short: "Authenticate and unlock your keeper",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client := keeper.NewClient(apiURL)

			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println("")

			if err = client.Login(ctx, args[0], string(password)); err != nil {
				return fmt.Errorf("failed to login: %w", err)
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

	flags := cmd.Flags()
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the keeper api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")

	return cmd
}
