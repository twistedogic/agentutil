## 1. Core Package

- [x] 1.1 Create `tools/todos/todos.go` with `TodoItem`, `Response`, and `Store` types
- [x] 1.2 Implement `Store.Load()` — read `.todos.json`, return empty slice if file absent
- [x] 1.3 Implement `Store.Save()` — write full list to file as JSON
- [x] 1.4 Implement `Update(store, items)` — validate statuses, diff against old state, compute transitions and counts, save and return `Response`
- [x] 1.5 Implement `NewTodosTool(file string) fantasy.AgentTool` — wraps `Update` with fantasy tool interface

## 2. CLI Wiring

- [x] 2.1 Create root `todos.go` with `newTodoCmd()` cobra command
- [x] 2.2 Accept JSON array as first positional argument, unmarshal to `[]TodoItem`
- [x] 2.3 Add `--file` flag defaulting to `.todos.json`
- [x] 2.4 Register `newTodoCmd()` in `main.go`

## 3. Skill

- [x] 3.1 Create `skills/todos/SKILL.md` with description, usage examples, input schema, output schema, and guidance on when to use (keep exactly one `in_progress`, skip for simple tasks)

## 4. Tests

- [x] 4.1 Write unit tests for `Store.Load` / `Store.Save` using temp dirs
- [x] 4.2 Write unit tests for `Update` covering: first write, replacement, invalid status, just-started transition, just-completed transition, no-transition
- [x] 4.3 Run `go test ./tools/todos/...` and verify all pass

## 5. Verification

- [x] 5.1 Run `go build ./...` — no compile errors
- [x] 5.2 Run `go test ./...` — all tests pass
- [x] 5.3 Smoke-test CLI: `agentutil todo '[{"content":"test","status":"pending","active_form":"Testing"}]'`
- [x] 5.4 Smoke-test `--file` flag with a custom path
