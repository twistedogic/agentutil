package travel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

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

func nominatimResult(lat, lon, display string) []byte {
	result := []map[string]string{{"lat": lat, "lon": lon, "display_name": display}}
	b, _ := json.Marshal(result)
	return b
}

func osrmOkResult(duration, distance float64) []byte {
	body := map[string]any{
		"code": "Ok",
		"routes": []map[string]any{
			{"duration": duration, "distance": distance},
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func TestGeocodeSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nominatimResult("40.7128", "-74.0060", "New York, NY, USA"))
	}))
	defer srv.Close()

	c, err := geocode(context.Background(), testClient(srv), "New York")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.displayName != "New York, NY, USA" {
		t.Errorf("expected display name %q, got %q", "New York, NY, USA", c.displayName)
	}
	if c.lat == 0 || c.lon == 0 {
		t.Error("expected non-zero lat/lon")
	}
}

func TestGeocodeNoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	_, err := geocode(context.Background(), testClient(srv), "xyzzy404notfound")
	if err == nil {
		t.Fatal("expected error for no results, got nil")
	}
}

func TestSortByDuration(t *testing.T) {
	routes := []TravelOption{
		{Mode: "foot", DurationSeconds: 3600, DistanceMeters: 5000},
		{Mode: "driving", DurationSeconds: 600, DistanceMeters: 10000},
		{Mode: "cycling", DurationSeconds: 1200, DistanceMeters: 8000},
	}

	// Simulate the sort applied in osrmRoutes.
	sortRoutes(routes)

	if routes[0].Mode != "driving" {
		t.Errorf("expected driving first, got %s", routes[0].Mode)
	}
	if routes[1].Mode != "cycling" {
		t.Errorf("expected cycling second, got %s", routes[1].Mode)
	}
	if routes[2].Mode != "foot" {
		t.Errorf("expected foot last, got %s", routes[2].Mode)
	}
}

func TestNewTravelToolInfo(t *testing.T) {
	tool := NewTravelTool(nil)
	info := tool.Info()
	if info.Name != "travel" {
		t.Errorf("expected tool name 'travel', got %q", info.Name)
	}
}
