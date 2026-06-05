package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"charm.land/fantasy"
)

const toolDescription = "Manage a structured task list for multi-step work; each task has pending/in_progress/completed state. Keep exactly one task in_progress at a time. Skip for simple or single-step tasks."

type TodoItem struct {
	Content    string `json:"content" description:"What needs to be done (imperative form)"`
	Status     string `json:"status" description:"Task status: pending, in_progress, or completed"`
	ActiveForm string `json:"active_form" description:"Present continuous form (e.g., 'Running tests')"`
}

type Response struct {
	IsNew         bool       `json:"is_new"`
	Todos         []TodoItem `json:"todos"`
	JustCompleted []string   `json:"just_completed,omitempty"`
	JustStarted   string     `json:"just_started,omitempty"`
	Completed     int        `json:"completed"`
	Pending       int        `json:"pending"`
	InProgress    int        `json:"in_progress"`
	Total         int        `json:"total"`
}

type Store struct {
	Path string
}

func (s Store) Load() ([]TodoItem, error) {
	b, err := os.ReadFile(s.Path)
	if os.IsNotExist(err) {
		return []TodoItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read todos file: %w", err)
	}
	var items []TodoItem
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, fmt.Errorf("failed to parse todos file: %w", err)
	}
	return items, nil
}

func (s Store) Save(items []TodoItem) error {
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}
	if err := os.WriteFile(s.Path, b, 0644); err != nil {
		return fmt.Errorf("failed to write todos file: %w", err)
	}
	return nil
}

func Update(store Store, items []TodoItem) (Response, error) {
	for _, item := range items {
		switch item.Status {
		case "pending", "in_progress", "completed":
		default:
			return Response{}, fmt.Errorf("invalid status %q for todo %q", item.Status, item.Content)
		}
	}

	old, err := store.Load()
	if err != nil {
		return Response{}, err
	}

	isNew := len(old) == 0
	oldStatus := make(map[string]string, len(old))
	for _, o := range old {
		oldStatus[o.Content] = o.Status
	}

	var justCompleted []string
	var justStarted string
	completed, pending, inProgress := 0, 0, 0

	for _, item := range items {
		prev, existed := oldStatus[item.Content]
		switch item.Status {
		case "completed":
			completed++
			if existed && prev != "completed" {
				justCompleted = append(justCompleted, item.Content)
			}
		case "in_progress":
			inProgress++
			if !existed || prev != "in_progress" {
				if item.ActiveForm != "" {
					justStarted = item.ActiveForm
				} else {
					justStarted = item.Content
				}
			}
		case "pending":
			pending++
		}
	}

	if err := store.Save(items); err != nil {
		return Response{}, err
	}

	return Response{
		IsNew:         isNew,
		Todos:         items,
		JustCompleted: justCompleted,
		JustStarted:   justStarted,
		Completed:     completed,
		Pending:       pending,
		InProgress:    inProgress,
		Total:         len(items),
	}, nil
}

type params struct {
	Todos []TodoItem `json:"todos" description:"The updated todo list"`
}

func NewTodosTool(file string) fantasy.AgentTool {
	store := Store{Path: file}
	return fantasy.NewAgentTool(
		"todos",
		toolDescription,
		func(ctx context.Context, p params, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			resp, err := Update(store, p.Todos)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			b, err := json.Marshal(resp)
			if err != nil {
				return fantasy.NewTextErrorResponse("failed to encode response: " + err.Error()), nil
			}
			return fantasy.NewTextResponse(string(b)), nil
		},
	)
}
