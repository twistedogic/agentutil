# travel Specification

## Purpose
TBD - created by archiving change add-travel-tool. Update Purpose after archive.
## Requirements
### Requirement: Return travel options for two addresses
Given a source and destination address as free-text strings, the tool SHALL return all available travel modes (driving, cycling, foot) with their duration in seconds and distance in meters, sorted ascending by duration.

#### Scenario: Successful route lookup
- **WHEN** origin and destination are valid, resolvable addresses
- **THEN** the tool returns a JSON object with a `routes` array containing one entry per available mode, each with `mode`, `duration_seconds`, and `distance_meters`, sorted by `duration_seconds` ascending, and `origin_resolved` / `destination_resolved` fields showing the geocoded display names

#### Scenario: Mode with no route omitted
- **WHEN** OSRM returns no route for a given mode (e.g., no cycling path between continents)
- **THEN** that mode is silently omitted from the results; other modes are still returned

#### Scenario: Unresolvable address
- **WHEN** Nominatim cannot geocode the origin or destination
- **THEN** the tool returns an error message indicating which address failed to resolve

### Requirement: No-registration default backend
The tool SHALL function without any API key or environment variable configured, using Nominatim for geocoding and the OSRM demo server for routing.

#### Scenario: Default backend used when no env var set
- **WHEN** `ORS_API_KEY` is not set in the environment
- **THEN** the tool uses Nominatim + OSRM for all requests

### Requirement: OpenRouteService upgrade via env var
When `ORS_API_KEY` is set, the tool SHALL use OpenRouteService for both geocoding and routing instead of Nominatim + OSRM.

#### Scenario: ORS backend selected via env var
- **WHEN** `ORS_API_KEY` is set to a valid API key
- **THEN** the tool routes all requests through OpenRouteService and returns results in the same output format

#### Scenario: Invalid ORS API key
- **WHEN** `ORS_API_KEY` is set but the key is invalid or expired
- **THEN** the tool returns an error from the ORS API

### Requirement: Nominatim rate limiting with jitter
When using the default backend, the tool SHALL serialize Nominatim geocoding requests with a minimum 1-second delay plus random jitter between consecutive calls.

#### Scenario: Jitter applied between geocoding calls
- **WHEN** two Nominatim requests are made in sequence
- **THEN** a delay of at least 1 second (plus 0–500ms random jitter) elapses between the first and second request

