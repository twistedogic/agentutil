package todos

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) Store {
	t.Helper()
	return Store{Path: filepath.Join(t.TempDir(), ".todos.json")}
}

func TestStoreLoadMissing(t *testing.T) {
	s := tempStore(t)
	items, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty slice, got %d items", len(items))
	}
}

func TestStoreLoadSave(t *testing.T) {
	s := tempStore(t)
	in := []TodoItem{
		{Content: "do thing", Status: "pending", ActiveForm: "Doing thing"},
	}
	if err := s.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	if out[0].Content != in[0].Content {
		t.Errorf("content mismatch: got %q", out[0].Content)
	}
}

func TestStoreLoadBadJSON(t *testing.T) {
	s := Store{Path: filepath.Join(t.TempDir(), "bad.json")}
	if err := os.WriteFile(s.Path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load()
	if err == nil {
		t.Error("expected error for bad JSON, got nil")
	}
}

func TestUpdateInvalidStatus(t *testing.T) {
	s := tempStore(t)
	_, err := Update(s, []TodoItem{{Content: "x", Status: "unknown"}})
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestUpdateFirstWrite(t *testing.T) {
	s := tempStore(t)
	resp, err := Update(s, []TodoItem{
		{Content: "task", Status: "pending", ActiveForm: "Doing task"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsNew {
		t.Error("expected is_new=true on first write")
	}
	if resp.Total != 1 || resp.Pending != 1 {
		t.Errorf("unexpected counts: %+v", resp)
	}
}

func TestUpdateReplacement(t *testing.T) {
	s := tempStore(t)
	_, _ = Update(s, []TodoItem{
		{Content: "old", Status: "pending", ActiveForm: "Doing old"},
	})
	resp, err := Update(s, []TodoItem{
		{Content: "new", Status: "pending", ActiveForm: "Doing new"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsNew {
		t.Error("expected is_new=false on second write")
	}
	if resp.Total != 1 || resp.Todos[0].Content != "new" {
		t.Errorf("unexpected todos: %+v", resp.Todos)
	}
}

func TestUpdateJustStarted(t *testing.T) {
	s := tempStore(t)
	_, _ = Update(s, []TodoItem{
		{Content: "task", Status: "pending", ActiveForm: "Doing task"},
	})
	resp, err := Update(s, []TodoItem{
		{Content: "task", Status: "in_progress", ActiveForm: "Doing task"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.JustStarted != "Doing task" {
		t.Errorf("expected just_started=%q, got %q", "Doing task", resp.JustStarted)
	}
}

func TestUpdateJustCompleted(t *testing.T) {
	s := tempStore(t)
	_, _ = Update(s, []TodoItem{
		{Content: "task", Status: "in_progress", ActiveForm: "Doing task"},
	})
	resp, err := Update(s, []TodoItem{
		{Content: "task", Status: "completed", ActiveForm: "Doing task"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.JustCompleted) != 1 || resp.JustCompleted[0] != "task" {
		t.Errorf("expected just_completed=[task], got %v", resp.JustCompleted)
	}
}

func TestUpdateNoTransition(t *testing.T) {
	s := tempStore(t)
	items := []TodoItem{
		{Content: "task", Status: "in_progress", ActiveForm: "Doing task"},
	}
	_, _ = Update(s, items)
	resp, err := Update(s, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.JustStarted != "" {
		t.Errorf("expected empty just_started, got %q", resp.JustStarted)
	}
	if len(resp.JustCompleted) != 0 {
		t.Errorf("expected empty just_completed, got %v", resp.JustCompleted)
	}
}
