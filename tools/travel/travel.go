package travel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"charm.land/fantasy"
)

// TravelOption is one mode's result.
type TravelOption struct {
	Mode            string `json:"mode"`
	DurationSeconds int    `json:"duration_seconds"`
	DistanceMeters  int    `json:"distance_meters"`
}

// TravelResult is the full response returned by GetRoutes and NewTravelTool.
type TravelResult struct {
	OriginResolved      string         `json:"origin_resolved"`
	DestinationResolved string         `json:"destination_resolved"`
	Routes              []TravelOption `json:"routes"`
}

// coord holds a geocoded latitude/longitude pair with its display name.
type coord struct {
	lat, lon    float64
	displayName string
}

var osrmModes = []string{"driving", "cycling", "foot"}

// sortRoutes sorts routes ascending by DurationSeconds in place.
func sortRoutes(routes []TravelOption) {
	slices.SortFunc(routes, func(a, b TravelOption) int {
		return a.DurationSeconds - b.DurationSeconds
	})
}

// GetRoutes geocodes origin and destination then fetches travel options for all
// available modes, returning them sorted ascending by DurationSeconds.
// If ORS_API_KEY is set, OpenRouteService is used instead of Nominatim+OSRM.
func GetRoutes(ctx context.Context, client *http.Client, origin, destination string) (TravelResult, error) {
	if key := os.Getenv("ORS_API_KEY"); key != "" {
		return orsRoutes(ctx, client, origin, destination, key)
	}
	return osrmRoutes(ctx, client, origin, destination)
}

// osrmRoutes geocodes via Nominatim then queries the OSRM demo server.
func osrmRoutes(ctx context.Context, client *http.Client, origin, destination string) (TravelResult, error) {
	originCoord, err := geocode(ctx, client, origin)
	if err != nil {
		return TravelResult{}, fmt.Errorf("geocoding origin %q: %w", origin, err)
	}

	delay := time.Second + time.Duration(rand.IntN(500))*time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return TravelResult{}, ctx.Err()
	}

	destCoord, err := geocode(ctx, client, destination)
	if err != nil {
		return TravelResult{}, fmt.Errorf("geocoding destination %q: %w", destination, err)
	}

	type result struct {
		opt TravelOption
		err error
	}
	ch := make(chan result, len(osrmModes))
	var wg sync.WaitGroup
	for _, mode := range osrmModes {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			opt, err := osrmRoute(ctx, client, originCoord, destCoord, m)
			ch <- result{opt, err}
		}(mode)
	}
	wg.Wait()
	close(ch)

	var routes []TravelOption
	for r := range ch {
		if r.err != nil {
			continue
		}
		routes = append(routes, r.opt)
	}
	sortRoutes(routes)

	return TravelResult{
		OriginResolved:      originCoord.displayName,
		DestinationResolved: destCoord.displayName,
		Routes:              routes,
	}, nil
}

// geocode resolves a free-text address to a coord using the Nominatim API.
func geocode(ctx context.Context, client *http.Client, address string) (coord, error) {
	apiURL := "https://nominatim.openstreetmap.org/search?q=" + url.QueryEscape(address) + "&format=json&limit=1"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return coord{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "agentutil/1.0 (https://github.com/twistedogic/agentutil)")

	resp, err := client.Do(req)
	if err != nil {
		return coord{}, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return coord{}, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	var results []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return coord{}, fmt.Errorf("decode nominatim response: %w", err)
	}
	if len(results) == 0 {
		return coord{}, fmt.Errorf("no results for address %q", address)
	}

	var lat, lon float64
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &lat); err != nil {
		return coord{}, fmt.Errorf("parse lat: %w", err)
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &lon); err != nil {
		return coord{}, fmt.Errorf("parse lon: %w", err)
	}
	return coord{lat: lat, lon: lon, displayName: results[0].DisplayName}, nil
}

// osrmRoute queries the OSRM demo server for a single mode between two coords.
// Returns an error (which callers silently drop) when no route is found.
func osrmRoute(ctx context.Context, client *http.Client, origin, dest coord, mode string) (TravelOption, error) {
	apiURL := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/%s/%f,%f;%f,%f?overview=false",
		mode, origin.lon, origin.lat, dest.lon, dest.lat,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return TravelOption{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return TravelOption{}, err
	}
	defer resp.Body.Close()

	var body struct {
		Code   string `json:"code"`
		Routes []struct {
			Duration float64 `json:"duration"`
			Distance float64 `json:"distance"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return TravelOption{}, err
	}
	if body.Code != "Ok" || len(body.Routes) == 0 {
		return TravelOption{}, fmt.Errorf("no route for mode %s", mode)
	}
	return TravelOption{
		Mode:            mode,
		DurationSeconds: int(body.Routes[0].Duration),
		DistanceMeters:  int(body.Routes[0].Distance),
	}, nil
}

// orsRoutes uses OpenRouteService when ORS_API_KEY is set.
// ORS accepts text addresses directly, so no separate geocoding step is needed.
func orsRoutes(ctx context.Context, client *http.Client, origin, destination, apiKey string) (TravelResult, error) {
	orsProfiles := map[string]string{
		"driving": "driving-car",
		"cycling": "cycling-regular",
		"foot":    "foot-walking",
	}

	type result struct {
		opt TravelOption
		err error
	}
	ch := make(chan result, len(orsProfiles))
	var wg sync.WaitGroup

	for mode, profile := range orsProfiles {
		wg.Add(1)
		go func(m, p string) {
			defer wg.Done()
			opt, err := orsRoute(ctx, client, origin, destination, m, p, apiKey)
			ch <- result{opt, err}
		}(mode, profile)
	}
	wg.Wait()
	close(ch)

	var routes []TravelOption
	for r := range ch {
		if r.err != nil {
			continue
		}
		routes = append(routes, r.opt)
	}
	sortRoutes(routes)

	return TravelResult{
		OriginResolved:      origin,
		DestinationResolved: destination,
		Routes:              routes,
	}, nil
}

// orsRoute queries the ORS directions API for one profile using structured geocoding.
func orsRoute(ctx context.Context, client *http.Client, origin, destination, mode, profile, apiKey string) (TravelOption, error) {
	geocodeURL := "https://api.openrouteservice.org/geocode/search?api_key=" + url.QueryEscape(apiKey) + "&text="
	geocodeOne := func(address string) ([2]float64, string, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", geocodeURL+url.QueryEscape(address), nil)
		if err != nil {
			return [2]float64{}, "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return [2]float64{}, "", err
		}
		defer resp.Body.Close()
		var body struct {
			Features []struct {
				Geometry struct {
					Coordinates [2]float64 `json:"coordinates"`
				} `json:"geometry"`
				Properties struct {
					Label string `json:"label"`
				} `json:"properties"`
			} `json:"features"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return [2]float64{}, "", err
		}
		if len(body.Features) == 0 {
			return [2]float64{}, "", fmt.Errorf("no geocoding result for %q", address)
		}
		return body.Features[0].Geometry.Coordinates, body.Features[0].Properties.Label, nil
	}

	origCoords, _, err := geocodeOne(origin)
	if err != nil {
		return TravelOption{}, fmt.Errorf("geocoding origin: %w", err)
	}
	destCoords, _, err := geocodeOne(destination)
	if err != nil {
		return TravelOption{}, fmt.Errorf("geocoding destination: %w", err)
	}

	bodyBytes, err := json.Marshal(map[string]any{
		"coordinates": [][2]float64{origCoords, destCoords},
	})
	if err != nil {
		return TravelOption{}, err
	}

	apiURL := "https://api.openrouteservice.org/v2/directions/" + profile
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TravelOption{}, err
	}
	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TravelOption{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TravelOption{}, fmt.Errorf("ORS returned status %d for mode %s", resp.StatusCode, mode)
	}

	var orsResp struct {
		Routes []struct {
			Summary struct {
				Duration float64 `json:"duration"`
				Distance float64 `json:"distance"`
			} `json:"summary"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orsResp); err != nil {
		return TravelOption{}, err
	}
	if len(orsResp.Routes) == 0 {
		return TravelOption{}, fmt.Errorf("no route from ORS for mode %s", mode)
	}
	return TravelOption{
		Mode:            mode,
		DurationSeconds: int(orsResp.Routes[0].Summary.Duration),
		DistanceMeters:  int(orsResp.Routes[0].Summary.Distance),
	}, nil
}

// NewTravelTool creates a fantasy.AgentTool for travel time and distance lookup.
func NewTravelTool(client *http.Client) fantasy.AgentTool {
	if client == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxIdleConns = 100
		transport.MaxIdleConnsPerHost = 10
		transport.IdleConnTimeout = 90 * time.Second
		client = &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		}
	}

	return fantasy.NewAgentTool(
		"travel",
		"Get travel time and distance between two addresses. Returns available modes (driving, cycling, foot) sorted by duration ascending. Uses Nominatim+OSRM by default (no key required); set ORS_API_KEY to use OpenRouteService instead.",
		func(ctx context.Context, params travelParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Origin == "" {
				return fantasy.NewTextErrorResponse("origin parameter is required"), nil
			}
			if params.Destination == "" {
				return fantasy.NewTextErrorResponse("destination parameter is required"), nil
			}

			result, err := GetRoutes(ctx, client, params.Origin, params.Destination)
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

type travelParams struct {
	Origin      string `json:"origin" description:"The starting address or place name"`
	Destination string `json:"destination" description:"The destination address or place name"`
}
