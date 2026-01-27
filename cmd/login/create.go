package login

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/davidsbond/x/envvar"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/davidsbond/passwords/internal/cli/config"
	"github.com/davidsbond/passwords/pkg/passwords"
)

func create() *cobra.Command {
	var (
		apiURL     string
		configPath string
		domains    []string
	)

	cmd := &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new login",
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

			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println("")

			login := passwords.Login{
				Username: args[0],
				Password: string(password),
				Domains:  domains,
			}

			if err = client.CreateLogin(ctx, login); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&apiURL, "api-url", envvar.String("PASSWORDS_API_URL", "http://localhost:8080"), "base url of the passwords api")
	flags.StringVar(&configPath, "config", envvar.String("PASSWORDS_CONFIG", config.DefaultConfigPath()), "path to config file")
	flags.StringSliceVar(&domains, "domains", nil, "domains where this login can be used")

	return cmd
}
