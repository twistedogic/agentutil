package main

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/travel"
)

func newTravelCmd() *cobra.Command {
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "travel <origin> <destination>",
		Short: "Get travel time and distance between two addresses, sorted by duration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, destination := args[0], args[1]
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			result, err := travel.GetRoutes(ctx, http.DefaultClient, origin, destination)
			if err != nil {
				return err
			}

			return writeJSON(result)
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "HTTP request timeout")
	return cmd
}
