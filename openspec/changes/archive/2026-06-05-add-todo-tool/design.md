## Context

Agentutil is a stateless CLI toolkit exposing agent tools via cobra commands and `charm.land/fantasy` AgentTool interfaces. All existing tools (fetch, search, wiki, lsp) are pure request/response with no persistent state. This change adds the first stateful tool — a todo list backed by `.todos.json` in the working directory.

The reference implementation is `charmbracelet/crush`'s `todos.go`, which stores todo state in a session service. We replace that with a file-based store since agentutil has no session concept.

## Goals / Non-Goals

**Goals:**
- Persist todo state to `.todos.json` in cwd (configurable via `--file`)
- Match crush's diff logic: detect just-started and just-completed items across updates
- Expose as both a `fantasy.AgentTool` (for MCP/agent use) and a cobra CLI subcommand
- Ship a `skills/todos/SKILL.md` that agents can install and follow

**Non-Goals:**
- No concurrent access safety (single-agent, single-process assumption)
- No history or undo — each call replaces the full list
- No per-task metadata beyond content, status, and active_form

## Decisions

### File replaces session service
**Decision**: Use a flat JSON file (`[]TodoItem`) instead of a session service.
**Rationale**: Agentutil has no session infrastructure. A file in cwd is transparent, inspectable, and matches how agents operate — each tool invocation runs in the project directory where the file naturally lives.
**Alternative considered**: Pass full state as input parameter only (stateless) — rejected because it loses state between agent restarts and separate tool calls.

### Full list replacement (not patch)
**Decision**: Each call replaces the entire todo list, mirroring crush's design.
**Rationale**: Simplifies the API to a single array parameter. The agent always sends the complete desired state. Diff logic runs server-side to detect transitions.
**Alternative considered**: Add/update/remove individual items — rejected as more complex with no real benefit for the LLM use case.

### `NewTodosTool(file string)` — path at construction time
**Decision**: The file path is injected at tool construction, not exposed as a tool parameter.
**Rationale**: The LLM should not control which file state is written to. The path is an operator concern resolved at startup from cwd or `--file` flag.

### Package layout: `tools/todos/todos.go`
**Decision**: Mirror the existing `tools/fetch/`, `tools/search/` layout.
**Rationale**: Consistent with project conventions. The root `todos.go` file wires the cobra command, the package handles logic.

## Risks / Trade-offs

- **Concurrent writes**: Two agent processes writing simultaneously will corrupt the file → acceptable given single-agent assumption; no mitigation needed now
- **Stale state**: If `.todos.json` is left over from a previous session it will be loaded → agents should call the tool with a fresh list at session start; document this in the skill
- **File path coupling**: Default cwd means tests must either use temp dirs or mock the store → use a `Store` interface or constructor injection in tests

## Open Questions

- Should the CLI `todo` command also support a `get` subcommand (read current state without modifying)? → Not in scope for this change; agents can read `.todos.json` directly.
