---
name: travel
description: Use when finding travel time and distance between two locations, comparing transport modes, or answering questions about how long it takes to get somewhere. Triggered by requests to "how long does it take to get from X to Y", "travel time between", "distance from X to Y", "compare driving vs walking", or any task requiring routing or commute information.
license: MIT
compatibility: Requires agentutil CLI
metadata:
  author: agentutil
  version: "1.0"
---

# Travel

Get travel time and distance between two addresses for all available transport modes (driving, cycling, walking), sorted fastest first.

## When to Use

- Answering "how long does it take to get from A to B?"
- Comparing transport modes for a journey
- Estimating commute times or trip distances
- Any task requiring routing or travel logistics

## Tool Usage

```bash
agentutil travel <origin> <destination>
```

**Arguments:**
- `origin`: Starting address or place name (free text)
- `destination`: Destination address or place name (free text)

**Options:**
- `--timeout <duration>`: HTTP request timeout (default: `60s`)

## Output Format

```json
{
  "origin_resolved": "Times Square, Manhattan, New York, NY 10036, United States",
  "destination_resolved": "John F. Kennedy International Airport, Jamaica, Queens, NY, United States",
  "routes": [
    { "mode": "driving",  "duration_seconds": 2820, "distance_meters": 26300 },
    { "mode": "cycling",  "duration_seconds": 5400, "distance_meters": 22100 },
    { "mode": "foot",     "duration_seconds": 18000, "distance_meters": 19800 }
  ]
}
```

- `origin_resolved` / `destination_resolved`: Geocoded display names showing what was matched
- `routes`: All available modes sorted by `duration_seconds` ascending (fastest first)
- Modes with no available route (e.g., no cycling path) are silently omitted

## Backend

By default, uses **Nominatim** (OpenStreetMap geocoding) + **OSRM** demo server (routing). No API key or registration required.

For better reliability and higher quota, set `ORS_API_KEY` to use **OpenRouteService** (2,000 req/day free tier):

```bash
export ORS_API_KEY=your_key_here
agentutil travel "Paris, France" "Lyon, France"
```

Sign up at https://openrouteservice.org/dev/#/signup

## Rate Limits

| Backend | Limit | Notes |
|---------|-------|-------|
| Nominatim | 1 req/sec | Enforced with jitter between geocoding calls |
| OSRM demo | Unofficial | No SLA; may be slow or unavailable |
| OpenRouteService | 2,000 req/day | Requires free API key |

## Common Patterns

### Compare modes for a commute
```bash
agentutil travel "Brooklyn Bridge, New York" "Central Park, New York"
```

### International route
```bash
agentutil travel "Amsterdam, Netherlands" "Brussels, Belgium"
```

### Use ORS for reliability
```bash
ORS_API_KEY=xxx agentutil travel "10 Downing Street, London" "Heathrow Airport, London"
```

## Edge Cases

| Situation | Behavior |
|-----------|----------|
| Address not found by geocoder | Command exits non-zero with error indicating which address failed |
| Mode has no route (e.g., cycling across ocean) | Mode omitted from results; others still returned |
| OSRM demo unavailable | Command exits non-zero with HTTP error |
| Invalid `ORS_API_KEY` | Command exits non-zero with ORS API error |

## Installation

```bash
go install github.com/twistedogic/agentutil@latest
```

Verify:
```bash
agentutil travel --help
```
