package fetch

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractLinks(t *testing.T) {
	base, _ := url.Parse("https://example.com/page")

	tests := []struct {
		name     string
		html     string
		want     []string
		notWant  []string
	}{
		{
			name: "absolute href passthrough",
			html: `<a href="https://other.com/page">link</a>`,
			want: []string{"https://other.com/page"},
		},
		{
			name: "relative href resolved",
			html: `<a href="/about">about</a>`,
			want: []string{"https://example.com/about"},
		},
		{
			name: "mailto filtered",
			html: `<a href="mailto:x@y.com">email</a>`,
			notWant: []string{"mailto:x@y.com"},
		},
		{
			name: "javascript filtered",
			html: `<a href="javascript:void(0)">click</a>`,
			notWant: []string{"javascript:void(0)"},
		},
		{
			name: "fragment-only filtered",
			html: `<a href="#section">jump</a>`,
			notWant: []string{"#section"},
		},
		{
			name: "deduplication",
			html: `<a href="/foo">a</a><a href="/foo">b</a>`,
			want: []string{"https://example.com/foo"},
		},
		{
			name: "mixed",
			html: `<a href="/a">a</a><a href="mailto:x@y.com">b</a><a href="https://ext.com">c</a>`,
			want:    []string{"https://example.com/a", "https://ext.com"},
			notWant: []string{"mailto:x@y.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got := ExtractLinks(doc, base)
			gotSet := make(map[string]bool, len(got))
			for _, l := range got {
				gotSet[l] = true
			}
			for _, w := range tt.want {
				if !gotSet[w] {
					t.Errorf("expected %q in links, got %v", w, got)
				}
			}
			for _, nw := range tt.notWant {
				if gotSet[nw] {
					t.Errorf("did not expect %q in links, got %v", nw, got)
				}
			}
		})
	}
}
