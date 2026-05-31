package main

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/fetch"
)

func newFetchCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "fetch <url>",
		Short: "Fetch a URL and return content as markdown with extracted links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			content, links, err := fetch.FetchURLAndConvert(ctx, http.DefaultClient, url)
			if err != nil {
				return err
			}

			if links == nil {
				links = []string{}
			}

			return writeJSON(fetch.FetchResult{
				Content: content,
				Links:   links,
			})
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	return cmd
}
