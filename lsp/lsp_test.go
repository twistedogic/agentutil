package lsp

import (
	"context"
	"testing"
	"time"
)

// TestVersionedMapBasic tests basic VersionedMap operations.
func TestVersionedMapBasic(t *testing.T) {
	m := NewVersionedMap[string, int]()

	// Test Set and Get
	m.Set("key1", 100)
	v, ok := m.Get("key1")
	if !ok || v != 100 {
		t.Errorf("expected 100, got %d", v)
	}

	// Test Get non-existent
	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent key")
	}

	// Test Delete
	m.Del("key1")
	_, ok = m.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

// TestVersionedMapVersion tests that Version increments on writes.
func TestVersionedMapVersion(t *testing.T) {
	m := NewVersionedMap[string, int]()

	v0 := m.Version()
	m.Set("a", 1)
	v1 := m.Version()
	if v1 <= v0 {
		t.Errorf("version should increment on Set, got v0=%d v1=%d", v0, v1)
	}

	m.Set("b", 2)
	v2 := m.Version()
	if v2 <= v1 {
		t.Errorf("version should increment on second Set, got v1=%d v2=%d", v1, v2)
	}

	m.Del("a")
	v3 := m.Version()
	if v3 <= v2 {
		t.Errorf("version should increment on Del, got v2=%d v3=%d", v2, v3)
	}
}

// TestVersionedMapCopy tests that Copy returns an independent snapshot.
func TestVersionedMapCopy(t *testing.T) {
	m := NewVersionedMap[string, int]()
	m.Set("x", 10)
	m.Set("y", 20)

	snapshot := m.Copy()
	if len(snapshot) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snapshot))
	}

	// Modifying snapshot doesn't affect original
	snapshot["z"] = 30
	if _, exists := m.Seq2()["z"]; exists {
		t.Error("copy should be independent")
	}
}

// TestManagerMap tests the generic thread-safe Map.
func TestManagerMap(t *testing.T) {
	m := NewMap[string, *Client]()

	// Empty
	if m.Len() != 0 {
		t.Errorf("expected len 0, got %d", m.Len())
	}

	// Set/Get
	m.Set("key1", nil)
	if v, ok := m.Get("key1"); !ok || v != nil {
		t.Error("Set/Get failed")
	}

	// Len
	if m.Len() != 1 {
		t.Errorf("expected len 1, got %d", m.Len())
	}

	// Del
	m.Del("key1")
	if _, ok := m.Get("key1"); ok {
		t.Error("expected key1 to be deleted")
	}
}

// TestServerState tests ServerState constants.
func TestServerState(t *testing.T) {
	states := []ServerState{StateUnstarted, StateStarting, StateReady, StateError, StateStopped, StateDisabled}
	for _, s := range states {
		if int(s) < 0 {
			t.Errorf("invalid state: %v", s)
		}
	}
}

// TestDiagnosticCounts tests the DiagnosticCounts struct.
func TestDiagnosticCounts(t *testing.T) {
	counts := DiagnosticCounts{
		Error:       5,
		Warning:     3,
		Information: 2,
		Hint:        1,
	}
	if counts.Error != 5 || counts.Warning != 3 {
		t.Error("DiagnosticCounts mismatch")
	}
}

// TestHandlesFile tests file path matching.
func TestHandlesFile(t *testing.T) {
	cfg := ServerConfig{FileTypes: []string{".go", ".mod"}}
	client := &Client{
		name:      "gopls",
		cwd:       "/workspace",
		fileTypes: cfg.FileTypes,
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/workspace/main.go", true},
		{"/workspace/pkg/foo.go", true},
		{"/workspace/go.mod", true},
		{"/other/file.go", false},      // outside cwd
		{"/workspace/main.ts", false},   // wrong extension
	}

	for _, tt := range tests {
		if got := client.HandlesFile(tt.path); got != tt.expected {
			t.Errorf("HandlesFile(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

// TestRecentlyUnavailable tests unavailable tracking.
func TestRecentlyUnavailable(t *testing.T) {
	m := &Manager{
		clients:     NewMap[string, *Client](),
		unavailable: NewMap[string, time.Time](),
		now:         time.Now,
	}

	// Not unavailable initially
	if m.recentlyUnavailable("gopls") {
		t.Error("should not be unavailable initially")
	}

	// Mark as unavailable
	m.markUnavailable("gopls")

	// Should be unavailable now
	if !m.recentlyUnavailable("gopls") {
		t.Error("should be unavailable after mark")
	}

	// Time travel past the retry delay
	m.unavailable.Set("gopls", time.Now().Add(-unavailableRetryDelay - 1))
	if m.recentlyUnavailable("gopls") {
		t.Error("should not be unavailable after retry delay")
	}
}

// TestManagerStart tests that Manager.Start doesn't panic on nil context.
func TestManagerStart(t *testing.T) {
	m := &Manager{
		clients:     NewMap[string, *Client](),
		unavailable: NewMap[string, time.Time](),
		now:         time.Now,
	}

	// Should not panic
	ctx := context.Background()
	m.Start(ctx, "/workspace/main.go")
}