package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"charm.land/fantasy"
	"github.com/twistedogic/agentutil/tools/fetch"
)

// WikiResult is the structured response returned by WikiSearch and NewWikiTool.
type WikiResult struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Links   []string `json:"links"`
}

const openSearchAPI = "https://en.wikipedia.org/w/api.php"

// WikiSearch queries the Wikipedia OpenSearch API for query, fetches the top
// result's article content, and returns a WikiResult.
func WikiSearch(ctx context.Context, client *http.Client, query string) (WikiResult, error) {
	apiURL := openSearchAPI + "?action=opensearch&search=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return WikiResult{}, fmt.Errorf("failed to create opensearch request: %w", err)
	}
	req.Header.Set("User-Agent", "curl/8.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return WikiResult{}, fmt.Errorf("opensearch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WikiResult{}, fmt.Errorf("opensearch request returned status %d", resp.StatusCode)
	}

	// OpenSearch response: [query, titles[], descriptions[], urls[]]
	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return WikiResult{}, fmt.Errorf("failed to parse opensearch response: %w", err)
	}
	if len(raw) < 4 {
		return WikiResult{}, fmt.Errorf("unexpected opensearch response format")
	}

	var titles []string
	var urls []string
	if err := json.Unmarshal(raw[1], &titles); err != nil {
		return WikiResult{}, fmt.Errorf("failed to parse opensearch titles: %w", err)
	}
	if err := json.Unmarshal(raw[3], &urls); err != nil {
		return WikiResult{}, fmt.Errorf("failed to parse opensearch urls: %w", err)
	}

	if len(titles) == 0 || len(urls) == 0 {
		return WikiResult{}, fmt.Errorf("no Wikipedia results for %q", query)
	}

	title := titles[0]
	articleURL := urls[0]

	content, links, err := fetch.FetchURLAndConvert(ctx, client, articleURL)
	if err != nil {
		return WikiResult{}, fmt.Errorf("failed to fetch article: %w", err)
	}
	if links == nil {
		links = []string{}
	}

	return WikiResult{
		Title:   title,
		URL:     articleURL,
		Content: content,
		Links:   links,
	}, nil
}

// NewWikiTool creates a fantasy.AgentTool that searches Wikipedia and returns
// the top article as structured JSON.
func NewWikiTool(client *http.Client) fantasy.AgentTool {
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
		"wiki",
		"Search Wikipedia by topic and return the top article as clean markdown. Returns the article title, canonical URL, content, and all extracted links.",
		func(ctx context.Context, params wikiParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("query parameter is required"), nil
			}

			result, err := WikiSearch(ctx, client, params.Query)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			out, err := json.Marshal(result)
			if err != nil {
				return fantasy.NewTextErrorResponse("failed to encode result: " + err.Error()), nil
			}

			return fantasy.NewTextResponse(string(out)), nil
		},
	)
}

type wikiParams struct {
	Query string `json:"query" description:"The search term to look up on Wikipedia"`
}
