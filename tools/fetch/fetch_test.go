package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"charm.land/fantasy"
)

func invokeFetch(t *testing.T, tool fantasy.AgentTool, rawURL string) string {
	t.Helper()
	input, _ := json.Marshal(map[string]string{"url": rawURL})
	resp, err := tool.Run(context.Background(), fantasy.ToolCall{
		ID:    "test",
		Input: string(input),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return resp.Content
}

func TestNewFetchTool(t *testing.T) {
	t.Run("success HTML page with links", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><p>Hello</p><a href="/about">About</a></body></html>`)
		}))
		defer srv.Close()

		tool := NewFetchTool(srv.Client())
		raw := invokeFetch(t, tool, srv.URL)

		var result FetchResult
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			t.Fatalf("expected FetchResult JSON, got: %s", raw)
		}
		if result.Content == "" {
			t.Error("expected non-empty content")
		}
		if len(result.Links) == 0 {
			t.Error("expected at least one link")
		}
	})

	t.Run("non-200 status returns error text", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		tool := NewFetchTool(srv.Client())
		raw := invokeFetch(t, tool, srv.URL)

		// Should be an error string, not a valid FetchResult
		var result FetchResult
		if err := json.Unmarshal([]byte(raw), &result); err == nil && result.Content != "" {
			t.Errorf("expected error response, got content: %s", result.Content)
		}
	})

	t.Run("empty URL returns error", func(t *testing.T) {
		tool := NewFetchTool(nil)
		raw := invokeFetch(t, tool, "")
		if raw == "" {
			t.Error("expected error response")
		}
	})

	t.Run("invalid scheme returns error", func(t *testing.T) {
		tool := NewFetchTool(nil)
		raw := invokeFetch(t, tool, "ftp://example.com/file")
		if raw == "" {
			t.Error("expected error response")
		}
	})

	t.Run("nil client creates default tool", func(t *testing.T) {
		tool := NewFetchTool(nil)
		if tool == nil {
			t.Error("expected non-nil tool")
		}
	})
}
