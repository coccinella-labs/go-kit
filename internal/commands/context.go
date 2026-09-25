package commands

import (
	"fmt"

	"github.com/coccinella-labs/go-kit/internal/git"
	"github.com/spf13/cobra"
)

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Show repository context",
		Long:  `Display the current Git repository context.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := git.GetContext()
			if err != nil {
				return fmt.Errorf("failed to get repository context: %w", err)
			}

			cmd.Printf("Repository:\n")
			cmd.Printf("  owner: %s\n", ctx.Owner)
			cmd.Printf("  repo: %s\n", ctx.Repo)
			cmd.Printf("  branch: %s\n", ctx.Branch)
			cmd.Printf("  remote: %s\n", ctx.Remote)

			return nil
		},
	}

	return cmd
}
