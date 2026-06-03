package search

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"charm.land/fantasy"
)

// SearchResponse is the structured response returned by SearchTool.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// NewSearchTool creates a fantasy.AgentTool that searches DuckDuckGo and
// returns structured results as JSON.
func NewSearchTool(client *http.Client) fantasy.AgentTool {
	if client == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxIdleConns = 100
		transport.MaxIdleConnsPerHost = 10
		transport.IdleConnTimeout = 90 * time.Second
		client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	}

	return fantasy.NewAgentTool(
		"search",
		"Search the web via DuckDuckGo. Returns a JSON object with a 'results' array, each entry containing position, title, url, and snippet.",
		func(ctx context.Context, params searchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("query parameter is required"), nil
			}
			max := params.MaxResults
			if max <= 0 {
				max = 10
			}

			results, err := Search(ctx, client, params.Query, max)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			if results == nil {
				results = []SearchResult{}
			}

			out, err := json.Marshal(SearchResponse{Results: results})
			if err != nil {
				return fantasy.NewTextErrorResponse("failed to encode result: " + err.Error()), nil
			}

			return fantasy.NewTextResponse(string(out)), nil
		},
	)
}

type searchParams struct {
	Query      string `json:"query" description:"The search query"`
	MaxResults int    `json:"max_results,omitempty" description:"Maximum number of results to return (default 10)"`
}
