## Why

Agents often need to reason about travel logistics — how long it takes to get somewhere, how far, and by what means. There's currently no tool for this. Adding a travel tool gives agents actionable routing data using free, zero-registration APIs.

## What Changes

- New `travel` CLI command and `agentutil travel` tool
- Geocodes source and destination addresses via Nominatim (OpenStreetMap)
- Queries OSRM demo server for driving, cycling, and walking routes in parallel
- Returns all three modes sorted by ascending duration
- Falls back to OpenRouteService if `ORS_API_KEY` env var is set
- Nominatim calls are serialized with 1s base delay + random jitter to respect rate limits

## Capabilities

### New Capabilities

- `travel`: Given a source and destination address, return travel options (driving, cycling, foot) sorted by duration, with duration in seconds and distance in meters per mode.

### Modified Capabilities

<!-- none -->

## Impact

- New package `tools/travel/`
- New top-level `travel.go` (CLI command wiring, matching `wiki.go`/`search.go` pattern)
- `main.go`: register `newTravelCmd()`
- No changes to existing tools or packages
- New external dependencies: none (uses standard `net/http`)
- Runtime dependencies: Nominatim API, OSRM demo API (no key), or OpenRouteService (optional key via `ORS_API_KEY`)
