## ADDED Requirements

### Requirement: Update todo list
The system SHALL accept a full replacement list of todo items and persist it to a JSON file in the current working directory (`.todos.json` by default, overridable via `--file` flag).

#### Scenario: First write creates file
- **WHEN** the todo tool is called and no state file exists
- **THEN** the file is created with the provided items and `is_new` is `true` in the response

#### Scenario: Subsequent write replaces list
- **WHEN** the todo tool is called and a state file already exists
- **THEN** the file is overwritten with the new list and `is_new` is `false`

#### Scenario: Invalid status rejected
- **WHEN** an item has a status other than `pending`, `in_progress`, or `completed`
- **THEN** the tool returns an error and does not modify the file

### Requirement: Detect status transitions
The system SHALL compare the incoming list against the previously persisted list and report which items just started or just completed.

#### Scenario: Item transitions to in_progress
- **WHEN** an item's status changes from `pending` to `in_progress`
- **THEN** `just_started` in the response is set to that item's `active_form` (or `content` if `active_form` is empty)

#### Scenario: Item transitions to completed
- **WHEN** an item's status changes from `in_progress` or `pending` to `completed`
- **THEN** that item's `content` appears in `just_completed` in the response

#### Scenario: No transitions
- **WHEN** no item changes status from the previous state
- **THEN** `just_started` is empty and `just_completed` is empty

### Requirement: Return structured response
The system SHALL return a JSON response with the updated list and summary counts.

#### Scenario: Response shape
- **WHEN** the tool completes successfully
- **THEN** the response includes `todos`, `is_new`, `just_completed`, `just_started`, `completed`, `pending`, `in_progress`, and `total` fields

### Requirement: CLI subcommand
The system SHALL expose a `todo` subcommand in the agentutil CLI that accepts a JSON array of todo items as its first argument.

#### Scenario: Successful CLI invocation
- **WHEN** `agentutil todo '[{"content":"...","status":"pending","active_form":"..."}]'` is run
- **THEN** the response JSON is written to stdout

#### Scenario: Custom file path
- **WHEN** `agentutil todo --file /path/to/state.json '[...]'` is run
- **THEN** state is read from and written to the specified file instead of `.todos.json`

### Requirement: Agent skill
The system SHALL provide a `skills/todos/SKILL.md` file that describes when and how to use the todo tool, installable via `agentutil skills`.

#### Scenario: Skill installed
- **WHEN** `agentutil skills` is run
- **THEN** `skills/todos/SKILL.md` is copied to the target skills directory alongside other skills
