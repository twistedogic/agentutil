package wiki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func openSearchResponse(query string, titles, urls []string) []byte {
	resp := []any{query, titles, make([]string, len(titles)), urls}
	b, _ := json.Marshal(resp)
	return b
}

// rewriteTransport redirects all requests to a test server URL.
type rewriteTransport struct {
	serverURL string
	inner     http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(rt.serverURL)
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = target.Scheme
	cloned.URL.Host = target.Host
	cloned.Host = target.Host
	return rt.inner.RoundTrip(cloned)
}

func testClient(srv *httptest.Server) *http.Client {
	return &http.Client{
		Transport: &rewriteTransport{
			serverURL: srv.URL,
			inner:     http.DefaultTransport,
		},
	}
}

func TestWikiSearch_Success(t *testing.T) {
	articleFetched := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/w/api.php" {
			w.Header().Set("Content-Type", "application/json")
			w.Write(openSearchResponse("type theory",
				[]string{"Type theory"},
				[]string{"https://en.wikipedia.org/wiki/Type_theory"},
			))
			return
		}
		articleFetched = true
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Type theory</h1><p>A branch of mathematical logic.</p></body></html>`))
	}))
	defer srv.Close()

	result, err := WikiSearch(context.Background(), testClient(srv), "type theory")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Type theory" {
		t.Errorf("expected title 'Type theory', got %q", result.Title)
	}
	if result.URL == "" {
		t.Error("expected non-empty URL")
	}
	if result.Content == "" {
		t.Error("expected non-empty content")
	}
	if !articleFetched {
		t.Error("expected article fetch to be called")
	}
}

func TestWikiSearch_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(openSearchResponse("xyzzy404notfound", []string{}, []string{}))
	}))
	defer srv.Close()

	_, err := WikiSearch(context.Background(), testClient(srv), "xyzzy404notfound")
	if err == nil {
		t.Fatal("expected error for no results, got nil")
	}
}

func TestWikiSearch_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	_, err := WikiSearch(context.Background(), testClient(srv), "anything")
	if err == nil {
		t.Fatal("expected error for malformed response, got nil")
	}
}

func TestNewWikiTool_Info(t *testing.T) {
	tool := NewWikiTool(nil)
	info := tool.Info()
	if info.Name != "wiki" {
		t.Errorf("expected tool name 'wiki', got %q", info.Name)
	}
}
