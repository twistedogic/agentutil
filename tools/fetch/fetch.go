package fetch

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"charm.land/fantasy"
)

// FetchResult is the structured response returned by FetchTool.
type FetchResult struct {
	Content string   `json:"content"`
	Links   []string `json:"links"`
}

// NewFetchTool creates a fantasy.AgentTool that fetches a URL, converts the HTML
// response to markdown, and extracts absolute links from the page.
func NewFetchTool(client *http.Client) fantasy.AgentTool {
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
		"fetch",
		"Fetch raw content from a URL as markdown (max 5MB). Returns page content and all extracted absolute links. For HTML pages the content is converted to clean markdown with noisy elements removed.",
		func(ctx context.Context, params fetchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.URL == "" {
				return fantasy.NewTextErrorResponse("url parameter is required"), nil
			}
			lower := strings.ToLower(params.URL)
			if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
				return fantasy.NewTextErrorResponse("url must start with http:// or https://"), nil
			}

			content, links, err := FetchURLAndConvert(ctx, client, params.URL)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			if links == nil {
				links = []string{}
			}

			result := FetchResult{Content: content, Links: links}
			out, err := json.Marshal(result)
			if err != nil {
				return fantasy.NewTextErrorResponse("failed to encode result: " + err.Error()), nil
			}

			return fantasy.NewTextResponse(string(out)), nil
		},
	)
}

type fetchParams struct {
	URL string `json:"url" description:"The URL to fetch content from"`
}
