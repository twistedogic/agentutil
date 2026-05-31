package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const (
	BrowserUserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	maxFetchSize      = int64(5 * 1024 * 1024) // 5MB
)

var multipleNewlinesRe = regexp.MustCompile(`\n{3,}`)

var noisyTags = map[string]bool{
	"script": true, "style": true, "nav": true, "header": true,
	"footer": true, "aside": true, "noscript": true, "iframe": true, "svg": true,
}

// FetchURLAndConvert fetches a URL, converts HTML to markdown, and extracts absolute links.
func FetchURLAndConvert(ctx context.Context, client *http.Client, rawURL string) (content string, links []string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", BrowserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchSize))
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	raw := string(body)
	if !utf8.ValidString(raw) {
		return "", nil, fmt.Errorf("response content is not valid UTF-8")
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return raw, nil, nil
	}

	base, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	links = ExtractLinks(doc, base)

	cleaned := removeNoisyElements(raw)
	markdown, err := ConvertHTMLToMarkdown(cleaned)
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}
	content = cleanupMarkdown(markdown)

	return content, links, nil
}

// ExtractLinks collects href values from <a> elements, resolves them to absolute URLs,
// and filters out mailto:, javascript:, and fragment-only hrefs.
func ExtractLinks(doc *goquery.Document, base *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}
		lower := strings.ToLower(strings.TrimSpace(href))
		if strings.HasPrefix(lower, "mailto:") ||
			strings.HasPrefix(lower, "javascript:") ||
			strings.HasPrefix(lower, "#") {
			return
		}
		ref, err := url.Parse(href)
		if err != nil {
			return
		}
		abs := base.ResolveReference(ref).String()
		if !seen[abs] {
			seen[abs] = true
			links = append(links, abs)
		}
	})
	return links
}

// removeNoisyElements strips script, style, nav, header, footer, aside, noscript, iframe, svg nodes.
func removeNoisyElements(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}
	var remove func(*html.Node)
	remove = func(n *html.Node) {
		var toRemove []*html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && noisyTags[c.Data] {
				toRemove = append(toRemove, c)
			} else {
				remove(c)
			}
		}
		for _, node := range toRemove {
			n.RemoveChild(node)
		}
	}
	remove(doc)
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return htmlContent
	}
	return buf.String()
}

// ConvertHTMLToMarkdown converts HTML to markdown.
func ConvertHTMLToMarkdown(htmlContent string) (string, error) {
	converter := md.NewConverter("", true, nil)
	return converter.ConvertString(htmlContent)
}

// cleanupMarkdown collapses multiple blank lines and trims trailing whitespace.
func cleanupMarkdown(content string) string {
	content = multipleNewlinesRe.ReplaceAllString(content, "\n\n")
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
