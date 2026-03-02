package account

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func changePassword() *cobra.Command {
	return &cobra.Command{
		Use:   "change-password",
		Short: "Change your password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			oldPassword, err := cli.PromptPassword("Current password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			newPassword, err := cli.PromptPassword("New password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			if oldPassword == newPassword {
				return errors.New("new password cannot be the same as old password")
			}

			confirm, err := cli.PromptPassword("Confirm new password")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			if newPassword != confirm {
				return errors.New("new passwords do not match")
			}

			if err = client.ChangePassword(ctx, oldPassword, newPassword); err != nil {
				return fmt.Errorf("failed to change password: %w", err)
			}

			return nil
		},
	}
}
