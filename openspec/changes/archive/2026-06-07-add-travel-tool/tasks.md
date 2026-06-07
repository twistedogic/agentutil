## 1. Package scaffold

- [x] 1.1 Create `tools/travel/` directory and `travel.go` file with package declaration
- [x] 1.2 Define output types: `TravelOption` (mode, duration_seconds, distance_meters) and `TravelResult` (origin_resolved, destination_resolved, routes)

## 2. Geocoding

- [x] 2.1 Implement `geocode(ctx, client, address)` using Nominatim search API, returning lat/lon and display_name
- [x] 2.2 Add 1s + 0–500ms jitter delay between sequential Nominatim calls

## 3. OSRM routing

- [x] 3.1 Implement `osrmRoute(ctx, client, originCoord, destCoord, mode)` calling the OSRM demo route API and returning duration_seconds + distance_meters
- [x] 3.2 Fan out three parallel goroutines (driving, cycling, foot) using `sync.WaitGroup`, collect results, omit modes with no route

## 4. OpenRouteService backend

- [x] 4.1 Implement `orsRoute(ctx, client, origin, destination, apiKey)` calling ORS directions API for all three modes, returning the same output shape
- [x] 4.2 Read `ORS_API_KEY` env var in the tool constructor; if set, use ORS backend; otherwise use Nominatim + OSRM

## 5. Tool wiring

- [x] 5.1 Implement `NewTravelTool(client)` returning a `fantasy.AgentTool` with params `origin` and `destination`
- [x] 5.2 Sort results ascending by `duration_seconds` before returning
- [x] 5.3 Create `tools/travel/travel_test.go` with unit tests for geocode parsing and result sorting

## 6. CLI command

- [x] 6.1 Create top-level `travel.go` with `newTravelCmd()` matching the pattern of `wiki.go` / `search.go`
- [x] 6.2 Register `newTravelCmd()` in `main.go`

## 7. Skill documentation

- [x] 7.1 Create `skills/travel/SKILL.md` documenting usage, output format, env var, and rate limits
