package main

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/wiki"
)

func newWikiCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "wiki <query>",
		Short: "Search Wikipedia and return the top article as markdown with extracted links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			result, err := wiki.WikiSearch(ctx, http.DefaultClient, query)
			if err != nil {
				return err
			}

			return writeJSON(result)
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	return cmd
}
