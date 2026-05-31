package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceRoot(t *testing.T) {
	tmp := t.TempDir()

	// Create: tmp/project/go.mod, tmp/project/pkg/foo/foo.go
	projectDir := filepath.Join(tmp, "project")
	pkgDir := filepath.Join(projectDir, "pkg", "foo")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gomod := filepath.Join(projectDir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fooFile := filepath.Join(pkgDir, "foo.go")
	if err := os.WriteFile(fooFile, []byte("package foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("finds go.mod in ancestor", func(t *testing.T) {
		got := findWorkspaceRoot(fooFile)
		if got != projectDir {
			t.Errorf("got %q, want %q", got, projectDir)
		}
	})

	t.Run("returns file dir when no marker found", func(t *testing.T) {
		isolated := filepath.Join(tmp, "isolated")
		if err := os.MkdirAll(isolated, 0o755); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(isolated, "main.go")
		if err := os.WriteFile(f, []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := findWorkspaceRoot(f)
		if got != isolated {
			t.Errorf("got %q, want %q", got, isolated)
		}
	})

	t.Run("works with directory input", func(t *testing.T) {
		got := findWorkspaceRoot(pkgDir)
		if got != projectDir {
			t.Errorf("got %q, want %q", got, projectDir)
		}
	})
}

func TestExpandGlobs(t *testing.T) {
	tmp := t.TempDir()

	files := []string{"a.go", "b.go", "sub/c.go"}
	for _, f := range files {
		full := filepath.Join(tmp, f)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("single pattern", func(t *testing.T) {
		got, err := expandGlobs([]string{filepath.Join(tmp, "*.go")})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Errorf("want 2 files, got %d: %v", len(got), got)
		}
	})

	t.Run("doublestar recursive", func(t *testing.T) {
		got, err := expandGlobs([]string{filepath.Join(tmp, "**/*.go")})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Errorf("want 3 files, got %d: %v", len(got), got)
		}
	})

	t.Run("deduplicates overlapping patterns", func(t *testing.T) {
		got, err := expandGlobs([]string{
			filepath.Join(tmp, "*.go"),
			filepath.Join(tmp, "a.go"),
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Errorf("want 2 files after dedup, got %d: %v", len(got), got)
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		got, err := expandGlobs([]string{filepath.Join(tmp, "*.ts")})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("want 0 files, got %d", len(got))
		}
	})
}

func TestDiagnosticOutputJSON(t *testing.T) {
	d := DiagnosticOutput{
		File:     "/abs/path/foo.go",
		Line:     10,
		Col:      5,
		Severity: "error",
		Message:  "undefined: foo",
		Source:   "gopls",
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"file", "line", "col", "severity", "message", "source"} {
		if _, ok := got[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}

func TestRefOutputJSON(t *testing.T) {
	r := RefOutput{File: "/abs/path/bar.go", Line: 42, Col: 3}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"file", "line", "col"} {
		if _, ok := got[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}
