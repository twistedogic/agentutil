## ADDED Requirements

### Requirement: Search and fetch Wikipedia article
The system SHALL accept a free-text search query, resolve it to the top Wikipedia article using the OpenSearch API, fetch the article content, and return it as structured JSON containing the article title, canonical URL, markdown content, and extracted links.

#### Scenario: Successful search and fetch
- **WHEN** the agent invokes the wiki tool with a valid search query
- **THEN** the system returns a JSON object with non-empty `title`, `url`, `content`, and `links` fields representing the top Wikipedia article

#### Scenario: No results found
- **WHEN** the agent invokes the wiki tool with a query that returns no Wikipedia results
- **THEN** the system returns an error: `no Wikipedia results for "<query>"`

### Requirement: OpenSearch API integration
The system SHALL query `https://en.wikipedia.org/w/api.php?action=opensearch&search=<query>&limit=1` and parse the response to extract the first result's title and URL.

#### Scenario: Parse opensearch response
- **WHEN** the OpenSearch API returns a valid response with at least one result
- **THEN** the system extracts `titles[0]` and `urls[0]` from the response array

#### Scenario: API returns malformed response
- **WHEN** the OpenSearch API returns a response that cannot be parsed as a valid opensearch array
- **THEN** the system returns an error describing the parse failure

### Requirement: CLI subcommand
The system SHALL expose a `wiki` subcommand in the `agentutil` CLI that accepts a single positional argument (the search query) and writes the result as JSON to stdout.

#### Scenario: CLI invocation
- **WHEN** the user runs `agentutil wiki "<query>"`
- **THEN** the CLI prints a JSON object `{title, url, content, links}` to stdout

### Requirement: AgentTool integration
The system SHALL expose a `fantasy.AgentTool` named `wiki` with a `query` parameter that performs the same search-and-fetch operation and returns the result as a JSON string.

#### Scenario: AgentTool invocation
- **WHEN** an agent runtime calls the wiki tool with a `query` parameter
- **THEN** the tool returns a JSON text response with `title`, `url`, `content`, and `links`

#### Scenario: AgentTool missing query
- **WHEN** an agent runtime calls the wiki tool with an empty or missing `query` parameter
- **THEN** the tool returns a text error response: `query parameter is required`
