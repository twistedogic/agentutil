## Why

Agents working on multi-step tasks need a way to track and communicate their progress. A `todo` tool backed by a local `.todos.json` file gives agents lightweight persistent state without requiring a session server, and lets humans inspect or share task state across invocations.

## What Changes

- Add `tools/todos/` package with `Store`, `TodoItem`, `Response` types, file-based persistence, and `NewTodosTool`
- Add `agentutil todo` CLI subcommand accepting a JSON array of todos and a `--file` flag
- Add `skills/todos/SKILL.md` skill for agents

## Capabilities

### New Capabilities

- `todos`: Manage a structured todo list persisted to a local JSON file; supports full list replacement with diff-based metadata (just started, just completed, counts)

### Modified Capabilities

## Impact

- New `tools/todos/` package (no external dependencies beyond stdlib + `charm.land/fantasy`)
- New `todos.go` in root package wiring cobra command
- New `skills/todos/SKILL.md` embedded in the skills FS
- No changes to existing packages or APIs
