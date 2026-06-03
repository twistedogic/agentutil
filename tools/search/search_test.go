package search

import (
	"testing"
)

const sampleHTML = `<!DOCTYPE html>
<html>
<body>
<table>
<tr><td><a href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com&amp;rut=x" class="result-link">Example Domain</a></td></tr>
<tr><td class="result-snippet">This is the example snippet.</td></tr>
<tr><td><a href="https://golang.org" class="result-link">The Go Programming Language</a></td></tr>
<tr><td class="result-snippet">Go is an open source programming language.</td></tr>
</table>
</body>
</html>`

func TestParseLiteSearchResults(t *testing.T) {
	results, err := parseLiteSearchResults(sampleHTML, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Title != "Example Domain" {
		t.Errorf("expected title %q, got %q", "Example Domain", results[0].Title)
	}
	if results[0].URL != "https://example.com" {
		t.Errorf("expected url %q, got %q", "https://example.com", results[0].URL)
	}
	if results[0].Snippet != "This is the example snippet." {
		t.Errorf("expected snippet %q, got %q", "This is the example snippet.", results[0].Snippet)
	}
	if results[0].Position != 1 {
		t.Errorf("expected position 1, got %d", results[0].Position)
	}

	if results[1].URL != "https://golang.org" {
		t.Errorf("expected url %q, got %q", "https://golang.org", results[1].URL)
	}
	if results[1].Position != 2 {
		t.Errorf("expected position 2, got %d", results[1].Position)
	}
}

func TestParseLiteSearchResultsMaxResults(t *testing.T) {
	results, err := parseLiteSearchResults(sampleHTML, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (maxResults=1), got %d", len(results))
	}
}

func TestParseLiteSearchResultsEmpty(t *testing.T) {
	results, err := parseLiteSearchResults("<html><body></body></html>", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCleanDuckDuckGoURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com&rut=x", "https://example.com"},
		{"https://golang.org", "https://golang.org"},
		{"//duckduckgo.com/l/?uddg=https%3A%2F%2Fgolang.org", "https://golang.org"},
	}
	for _, c := range cases {
		got := cleanDuckDuckGoURL(c.in)
		if got != c.want {
			t.Errorf("cleanDuckDuckGoURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
