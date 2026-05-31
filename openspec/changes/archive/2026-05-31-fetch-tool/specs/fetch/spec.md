## ADDED Requirements

### Requirement: Fetch URL content as markdown
The `FetchTool` SHALL accept a URL parameter, perform an HTTP GET request, convert the HTML response body to clean markdown (stripping script, style, nav, header, footer, aside, noscript, iframe, and svg elements), and return the result as a string.

#### Scenario: Successful HTML fetch
- **WHEN** a valid `https://` or `http://` URL is provided
- **THEN** the tool returns a `FetchResult` with `Content` containing the page body as markdown and `Links` as a list of absolute URLs

#### Scenario: Non-HTML content type
- **WHEN** the response Content-Type is not `text/html`
- **THEN** the tool returns the raw response body as `Content` and an empty `Links` slice

#### Scenario: HTTP error status
- **WHEN** the server responds with a non-200 status code
- **THEN** the tool returns an error response with the status code

#### Scenario: Invalid URL scheme
- **WHEN** the URL does not start with `http://` or `https://`
- **THEN** the tool returns an error response without making a network request

#### Scenario: Empty URL
- **WHEN** the `url` parameter is empty
- **THEN** the tool returns an error response indicating the parameter is required

### Requirement: Extract absolute links from fetched page
The `FetchTool` SHALL extract all `href` attribute values from `<a>` elements in the HTML response and resolve them to absolute URLs using the request URL as the base.

#### Scenario: Relative href resolution
- **WHEN** the page contains an anchor with `href="/about"`  and the request URL is `https://example.com`
- **THEN** `Links` SHALL contain `https://example.com/about`

#### Scenario: Absolute href passthrough
- **WHEN** the page contains an anchor with `href="https://other.com/page"`
- **THEN** `Links` SHALL contain `https://other.com/page` unchanged

#### Scenario: Non-navigable hrefs filtered
- **WHEN** the page contains anchors with `href="mailto:x@y.com"`, `href="javascript:void(0)"`, or `href="#section"`
- **THEN** those values SHALL NOT appear in `Links`

### Requirement: Browser User-Agent on requests
The `FetchTool` SHALL send a realistic browser `User-Agent` header on all HTTP requests to improve compatibility with sites that block non-browser clients.

#### Scenario: User-Agent header set
- **WHEN** any HTTP GET request is made
- **THEN** the `User-Agent` header SHALL be a non-empty browser-style string

### Requirement: Configurable HTTP client
The `FetchTool` constructor SHALL accept an optional `*http.Client`. If `nil` is passed, it SHALL create a default client with a 30-second timeout, connection pooling, and 90-second idle connection timeout.

#### Scenario: Nil client falls back to default
- **WHEN** `NewFetchTool(nil)` is called
- **THEN** the tool uses an internally created `*http.Client` with a 30-second timeout

#### Scenario: Custom client used when provided
- **WHEN** a non-nil `*http.Client` is provided to `NewFetchTool`
- **THEN** all requests use that client
