package main

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/search"
)

func newSearchCmd() *cobra.Command {
	var timeout time.Duration
	var maxResults int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the web via DuckDuckGo and return structured JSON results",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			results, err := search.Search(ctx, http.DefaultClient, query, maxResults)
			if err != nil {
				return err
			}

			if results == nil {
				results = []search.SearchResult{}
			}

			return writeJSON(search.SearchResponse{Results: results})
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	cmd.Flags().IntVarP(&maxResults, "max", "n", 10, "Maximum number of results to return")
	return cmd
}
