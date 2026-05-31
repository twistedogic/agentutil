---
name: wiki
description: Use when looking up a topic on Wikipedia, getting an encyclopedia article, researching a concept, or retrieving reference information. Triggered by requests to "look up X on Wikipedia", "what is X", "search Wikipedia for", "get the Wikipedia article on", or any task requiring encyclopedic reference content.
license: MIT
compatibility: Requires agentutil CLI
metadata:
  author: agentutil
  version: "1.0"
---

# Wiki

Search Wikipedia by topic and return the top article as clean markdown, including the canonical URL and all extracted links.

## When to Use

- Looking up definitions, concepts, or background on a topic
- Getting a structured overview of a subject before deeper research
- Retrieving links to related Wikipedia articles for further exploration
- Any task needing encyclopedic reference content

## Tool Usage

```bash
agentutil wiki <query>
```

**Arguments:**
- `query`: The search term to look up on Wikipedia

**Options:**
- `--timeout <duration>`: HTTP request timeout (default: `30s`)

## Output Format

```json
{
  "title": "Type theory",
  "url": "https://en.wikipedia.org/wiki/Type_theory",
  "content": "# Type theory\n\nType theory is a branch of mathematical logic...",
  "links": [
    "https://en.wikipedia.org/wiki/Mathematical_logic",
    "https://en.wikipedia.org/wiki/Bertrand_Russell"
  ]
}
```

- `title`: The Wikipedia article title
- `url`: The canonical Wikipedia URL for the article
- `content`: Full article body converted to clean markdown
- `links`: All absolute URLs extracted from the article

## Common Patterns

### Look up a concept
```bash
agentutil wiki "type theory"
```

### Research a person
```bash
agentutil wiki "Ada Lovelace"
```

### Get background on a technology
```bash
agentutil wiki "WebAssembly"
```

## Edge Cases

| Situation | Behavior |
|-----------|----------|
| No Wikipedia results found | Command exits non-zero with error: `no Wikipedia results for "<query>"` |
| Ambiguous query | Returns the top OpenSearch result (may not be the intended article — refine query if needed) |
| Article > 5MB | Content truncated at 5MB |

## Installation

```bash
go install github.com/twistedogic/agentutil@latest
```

Verify:
```bash
agentutil wiki --help
```
