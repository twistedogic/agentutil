## ADDED Requirements

### Requirement: Fetch URL via CLI
The `agentutil fetch` command SHALL accept a single URL positional argument, fetch it using the `tools/fetch` package, and print a JSON object with `content` (markdown string) and `links` (array of absolute URL strings) to stdout.

#### Scenario: Successful fetch
- **WHEN** `agentutil fetch https://example.com` is invoked with a valid URL
- **THEN** the command exits 0 and prints a JSON object `{"content": "...", "links": [...]}` to stdout

#### Scenario: Missing URL argument
- **WHEN** `agentutil fetch` is invoked with no arguments
- **THEN** the command exits non-zero and prints a usage error

#### Scenario: Invalid URL scheme
- **WHEN** `agentutil fetch ftp://example.com` is invoked with a non-http/https URL
- **THEN** the command exits non-zero and prints a JSON error object to stderr

#### Scenario: Network or HTTP error
- **WHEN** the URL is unreachable or returns a non-200 status
- **THEN** the command exits non-zero and prints a JSON error object to stderr

### Requirement: Configurable fetch timeout
The `agentutil fetch` command SHALL accept a `--timeout` flag (default `30s`) that controls the HTTP request timeout.

#### Scenario: Default timeout
- **WHEN** `agentutil fetch <url>` is invoked without `--timeout`
- **THEN** the HTTP client uses a 30-second timeout

#### Scenario: Custom timeout
- **WHEN** `agentutil fetch <url> --timeout 10s` is invoked
- **THEN** the HTTP client uses a 10-second timeout
