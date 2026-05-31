package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/lsp"
)

// rootMarkers are used to detect workspace root.
var rootMarkers = []string{
	"go.mod", "package.json", "Cargo.toml",
	"pyproject.toml", "requirements.txt", ".git",
}

// findWorkspaceRoot walks up from dir until a root marker is found.
func findWorkspaceRoot(path string) string {
	dir := path
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	dir, _ = filepath.Abs(dir)
	for {
		for _, marker := range rootMarkers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return filepath.Dir(path)
	}
	return path
}

// noopResolver is a VariableResolver that returns values unchanged.
type noopResolver struct{}

func (noopResolver) ResolveValue(v string) (string, error) { return v, nil }

// cliConfigStore satisfies lsp.ConfigStore for CLI use.
type cliConfigStore struct {
	serverOverride string
}

func (s *cliConfigStore) LSP() map[string]lsp.ServerConfig {
	if s.serverOverride == "" {
		return nil
	}
	return map[string]lsp.ServerConfig{
		s.serverOverride: {Command: s.serverOverride},
	}
}

func (s *cliConfigStore) AutoLSP() *bool {
	v := true
	if s.serverOverride != "" {
		v = false
	}
	return &v
}

func (s *cliConfigStore) Resolver() lsp.VariableResolver { return noopResolver{} }

// writeError writes a JSON error object to stderr.
func writeError(err error) {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Fprintln(os.Stderr, string(b))
}

// expandGlobs expands glob patterns and returns deduplicated absolute paths.
func expandGlobs(patterns []string) ([]string, error) {
	seen := map[string]bool{}
	var result []string
	for _, pattern := range patterns {
		absPattern, err := filepath.Abs(pattern)
		if err != nil {
			return nil, err
		}
		matches, err := doublestar.FilepathGlob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				result = append(result, m)
			}
		}
	}
	return result, nil
}

// severityName converts LSP DiagnosticSeverity to a string.
func severityName(s int) string {
	switch s {
	case 1:
		return "error"
	case 2:
		return "warning"
	case 3:
		return "information"
	case 4:
		return "hint"
	default:
		return "unknown"
	}
}

// DiagnosticOutput is the JSON output shape for a single diagnostic.
type DiagnosticOutput struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source"`
}

// RefOutput is the JSON output shape for a single reference location.
type RefOutput struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

func newLSPCmd() *cobra.Command {
	var server string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "LSP-powered code analysis",
	}
	cmd.PersistentFlags().StringVar(&server, "server", "", "Force a specific LSP server (e.g. gopls)")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "Timeout for LSP startup and settling")

	cmd.AddCommand(newDiagnosticsCmd(&server, &timeout))
	cmd.AddCommand(newRefsCmd(&server, &timeout))
	return cmd
}

func newDiagnosticsCmd(server *string, timeout *time.Duration) *cobra.Command {
	return &cobra.Command{
		Use:   "diagnostics <pattern> [pattern...]",
		Short: "Get LSP diagnostics for files matching glob patterns",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := expandGlobs(args)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Println("[]")
				return nil
			}

			root := findWorkspaceRoot(files[0])
			store := &cliConfigStore{serverOverride: *server}
			mgr := lsp.NewManager(store, root)

			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			mgr.Start(ctx, files[0])

			clients := mgr.Clients()
			var activeClients []*lsp.Client
			for name := range clients.Seq() {
				if c, ok := clients.Get(name); ok {
					activeClients = append(activeClients, c)
				}
			}

			for _, c := range activeClients {
				for _, f := range files {
					if c.HandlesFile(f) {
						if err := c.OpenFile(ctx, f); err != nil {
							return fmt.Errorf("opening %s: %w", f, err)
						}
					}
				}
			}

			for _, c := range activeClients {
				c.WaitForDiagnostics(ctx, *timeout)
			}

			var out []DiagnosticOutput
			for _, c := range activeClients {
				diags := c.GetDiagnostics()
				for uri, ds := range diags {
					filePath := uriToPath(string(uri))
					for _, d := range ds {
						out = append(out, DiagnosticOutput{
							File:     filePath,
							Line:     int(d.Range.Start.Line) + 1,
							Col:      int(d.Range.Start.Character) + 1,
							Severity: severityName(int(d.Severity)),
							Message:  d.Message,
							Source:   c.GetName(),
						})
					}
				}
			}

			if out == nil {
				out = []DiagnosticOutput{}
			}
			return writeJSON(out)
		},
	}
}

func newRefsCmd(server *string, timeout *time.Duration) *cobra.Command {
	return &cobra.Command{
		Use:   "refs <file> <line> <col>",
		Short: "Find all references to the symbol at file:line:col",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			line, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid line %q: %w", args[1], err)
			}
			col, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("invalid col %q: %w", args[2], err)
			}

			root := findWorkspaceRoot(filePath)
			store := &cliConfigStore{serverOverride: *server}
			mgr := lsp.NewManager(store, root)

			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			defer cancel()

			mgr.Start(ctx, filePath)

			clients := mgr.Clients()
			var target *lsp.Client
			for name := range clients.Seq() {
				if c, ok := clients.Get(name); ok && c.HandlesFile(filePath) {
					target = c
					break
				}
			}
			if target == nil {
				return fmt.Errorf("no LSP server handles file %s", filePath)
			}

			if err := target.OpenFile(ctx, filePath); err != nil {
				return fmt.Errorf("opening %s: %w", filePath, err)
			}
			target.WaitForDiagnostics(ctx, *timeout)

			locs, err := target.FindReferences(ctx, filePath, line, col, false)
			if err != nil {
				return fmt.Errorf("find references: %w", err)
			}

			var out []RefOutput
			for _, loc := range locs {
				out = append(out, RefOutput{
					File: uriToPath(string(loc.URI)),
					Line: int(loc.Range.Start.Line) + 1,
					Col:  int(loc.Range.Start.Character) + 1,
				})
			}
			if out == nil {
				out = []RefOutput{}
			}
			return writeJSON(out)
		},
	}
}

// writeJSON marshals v and writes it to stdout.
func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// uriToPath converts a file URI to a local path.
func uriToPath(uri string) string {
	const prefix = "file://"
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		return uri[len(prefix):]
	}
	return uri
}
