## ADDED Requirements

### Requirement: Search returns structured JSON results
The system SHALL query DuckDuckGo lite and return a JSON object `{"results": [...]}` where each result contains `position` (int), `title` (string), `url` (string), and `snippet` (string).

#### Scenario: Successful search
- **WHEN** a valid query is provided
- **THEN** the command outputs a JSON object with a `results` array containing at least one entry with non-empty `title` and `url` fields

#### Scenario: No results
- **WHEN** the query returns no matches
- **THEN** the command outputs `{"results": []}` and exits with code 0

### Requirement: Search CLI flags
The `search` command SHALL accept `--timeout` (duration, default 30s) and `--max` / `-n` (int, default 10) flags controlling request timeout and maximum result count.

#### Scenario: Custom max results
- **WHEN** `--max 5` is passed
- **THEN** the results array contains at most 5 entries

#### Scenario: Custom timeout
- **WHEN** `--timeout 5s` is passed and the request completes within 5s
- **THEN** the command succeeds normally

#### Scenario: Timeout exceeded
- **WHEN** `--timeout 1ms` is passed (effectively zero)
- **THEN** the command exits non-zero with an error message

### Requirement: Pre-request delay
The system SHALL sleep a random duration between 500ms and 2000ms before issuing the HTTP request on every invocation.

#### Scenario: Delay is applied
- **WHEN** a search is executed
- **THEN** at least 500ms elapses before the HTTP request is sent

### Requirement: Fantasy agent tool integration
The `tools/search` package SHALL expose a `NewSearchTool(client *http.Client) fantasy.AgentTool` constructor following the same pattern as `tools/fetch`.

#### Scenario: Tool registration
- **WHEN** `NewSearchTool` is called with a nil client
- **THEN** a default HTTP client is created and the tool is returned without error
