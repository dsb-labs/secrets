package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/cmd/account"
	"github.com/dsb-labs/secrets/cmd/auth"
	"github.com/dsb-labs/secrets/cmd/login"
	"github.com/dsb-labs/secrets/cmd/note"
	"github.com/dsb-labs/secrets/cmd/serve"
	"github.com/dsb-labs/secrets/cmd/tool"
)

//go:generate go tool mockery
//go:generate go tool templ generate -include-version=false
//go:generate go tool go-licenses save --one_output --force --ignore "github.com/dsb-labs/secrets,filippo.io/csrf" --save_path licenses .
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cmd := &cobra.Command{
		Use:          "secrets",
		Short:        "A simple, self-hostable secret manager",
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		cmd.Version = info.Main.Version
	}

	cmd.AddCommand(
		serve.Command(),
		auth.Command(),
		login.Command(),
		note.Command(),
		account.Command(),
		tool.Command(),
	)

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
