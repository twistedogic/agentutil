## Context

The codebase is a Go CLI (`agentutil`) that exposes agent-friendly tools: search, fetch, wiki, todos, LSP. Each tool lives in `tools/<name>/` and is wired into a top-level `<name>.go` CLI command. There is no existing routing or geolocation capability. External HTTP APIs are used directly via `net/http` — no SDK wrappers.

## Goals / Non-Goals

**Goals:**
- Given two free-text addresses, return travel options (driving, cycling, foot) sorted ascending by duration
- Zero mandatory configuration — works out of the box with no API key
- Optional upgrade path: set `ORS_API_KEY` to use OpenRouteService instead of OSRM+Nominatim
- Respect Nominatim's 1 req/s policy with jitter

**Non-Goals:**
- Turn-by-turn directions or route geometry
- Transit (bus/rail) routing
- Multi-stop/waypoint routing
- Real-time traffic data
- Caching or persisting results

## Decisions

### 1. Default backend: Nominatim + OSRM

**Decision**: Use Nominatim (geocoding) + OSRM demo server (routing) when no API key is present.

**Rationale**: Both are fully free, require no registration, and are usable immediately. OSRM is the reference implementation of road routing on OpenStreetMap data and returns reliable duration/distance values. Nominatim is the canonical OSM geocoder.

**Alternative considered**: Use only OpenRouteService (which accepts text addresses directly). Rejected because it requires an API key even for free tier, creating friction on first use.

### 2. ORS fallback via `ORS_API_KEY` env var

**Decision**: If `ORS_API_KEY` is set at runtime, switch to OpenRouteService for both geocoding and routing.

**Rationale**: ORS has a 2000 req/day free tier with better SLA than the unofficial OSRM demo server. Power users or production deployments can opt in without code changes. The env var pattern is already idiomatic in 12-factor apps and consistent with how API keys are handled in this project.

**Alternative considered**: A `--backend` flag. Rejected — env var is cleaner for persistent configuration and doesn't expose credentials in shell history.

### 3. Nominatim rate limiting: serialized calls with 1s + jitter

**Decision**: The two Nominatim geocoding requests are made sequentially. After the first, sleep `1s + rand(0, 500ms)` before the second.

**Rationale**: Nominatim's usage policy requires max 1 req/s. Jitter avoids synchronized bursts if multiple agent instances run simultaneously. The search tool uses the same jitter pattern (500–2000ms).

**Alternative considered**: Parallel geocoding calls. Rejected — violates Nominatim's rate limit and risks temporary IP bans.

### 4. OSRM routing calls: parallel across three modes

**Decision**: After geocoding, fire driving/cycling/foot OSRM requests concurrently via goroutines + `sync.WaitGroup`.

**Rationale**: The three OSRM calls are independent and hit different URL paths. Running them in parallel reduces total latency from ~3× to ~1× the single-call latency.

### 5. Output: all modes sorted by duration_seconds ascending

**Decision**: Return a JSON array of all three modes sorted by ascending `duration_seconds`. Each entry has `mode`, `duration_seconds`, `distance_meters`.

**Rationale**: Sorting by speed surfaces the quickest option first while giving the agent full context to reason about tradeoffs (e.g., cycling is only 5 min slower). A single-result response would lose that signal.

## Risks / Trade-offs

- **OSRM demo server reliability** → No SLA; could be unavailable. Mitigation: document clearly, encourage `ORS_API_KEY` for reliable use. HTTP errors are surfaced as tool errors, not panics.
- **Nominatim geocoding ambiguity** → "Springfield" could match many cities. Mitigation: return the `display_name` from Nominatim in each result so the agent can see what was resolved.
- **ORS free tier quota** → 2000 req/day. Each travel call costs 1 ORS route request (geocoding is built-in for ORS). Mitigation: documented in SKILL.md.
- **No result for some modes** → OSRM may return no route for cycling/foot over very long distances. Mitigation: omit modes that return no route rather than erroring.
