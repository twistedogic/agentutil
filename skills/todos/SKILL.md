---
name: todos
description: Manage a structured task list for multi-step work; each task has pending/in_progress/completed state. Keep exactly one task in_progress at a time. Skip for simple or single-step tasks.
license: MIT
compatibility: Requires agentutil CLI
metadata:
  author: agentutil
  version: "1.0"
---

# Todos

Maintain a structured task list for multi-step work, persisted to `.todos.json` in the current working directory.

## When to Use

- Working on a task with 3 or more distinct steps
- When you need to communicate progress to the user
- When state must survive across multiple tool calls or restarts

**Skip for**: Simple single-step tasks, quick questions, or one-off lookups.

## Tool Usage

```bash
agentutil todo update '<json-array>'   # replace the full todo list
agentutil todo list                    # read current todos
```

**Options (both subcommands):**
- `--file <path>`: Path to state file (default: `.todos.json` in cwd)

## Input Schema

```json
[
  {
    "content": "Implement the login handler",
    "status": "in_progress",
    "active_form": "Implementing the login handler"
  },
  {
    "content": "Write tests",
    "status": "pending",
    "active_form": "Writing tests"
  }
]
```

| Field | Type | Required | Description |
|---|---|---|---|
| `content` | string | yes | Task description in imperative form |
| `status` | string | yes | `pending`, `in_progress`, or `completed` |
| `active_form` | string | yes | Present continuous form shown during progress |

## Output Schema

```json
{
  "is_new": false,
  "todos": [...],
  "just_completed": ["Write tests"],
  "just_started": "Implementing the login handler",
  "completed": 1,
  "pending": 1,
  "in_progress": 1,
  "total": 3
}
```

## Rules

1. **Keep exactly one task `in_progress` at a time** — never mark two tasks as in_progress simultaneously
2. **Always send the full list** — each call replaces the entire todo list; omitting tasks removes them
3. **Update on transitions** — call the tool when starting a new task or completing one, not on every action
4. **Fresh list at session start** — if starting a new session on the same project, reset or replace the todo list to avoid stale state

## Common Patterns

### Start a multi-step task
```bash
agentutil todo update '[
  {"content":"Fetch data","status":"in_progress","active_form":"Fetching data"},
  {"content":"Process results","status":"pending","active_form":"Processing results"},
  {"content":"Write output","status":"pending","active_form":"Writing output"}
]'
```

### Mark a task complete and start the next
```bash
agentutil todo update '[
  {"content":"Fetch data","status":"completed","active_form":"Fetching data"},
  {"content":"Process results","status":"in_progress","active_form":"Processing results"},
  {"content":"Write output","status":"pending","active_form":"Writing output"}
]'
```

### Read current todos
```bash
agentutil todo list
```

### Custom file path
```bash
agentutil todo --file /tmp/my-session.json list
agentutil todo --file /tmp/my-session.json update '[...]'
```

## Installation

```bash
go install github.com/twistedogic/agentutil@latest
agentutil skills
```
