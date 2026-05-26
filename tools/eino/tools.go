// Package tools provides eino ADK tool adapters for lsplib.
package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/twistedogic/agentutil/lsp"
)

// DiagnosticsTool wraps LSP diagnostics as an eino Tool.
func DiagnosticsTool(manager *lsp.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"diagnostics",
		"Get LSP diagnostics (errors, warnings) for a file or all open files",
		func(ctx context.Context, params diagnosticsParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			filePath := params.FilePath
			// TODO: implement using manager
			return fantasy.NewTextResponse(
				fmt.Sprintf("diagnostics for: %s", filePath),
			), nil
		},
	)
}

type diagnosticsParams struct {
	FilePath string `json:"file_path"`
}

// ReferencesTool wraps LSP references as an eino Tool.
func ReferencesTool(manager *lsp.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"lsp_references",
		"Find all references to a symbol at a given position in a file",
		func(ctx context.Context, params referencesParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// TODO: implement using manager
			return fantasy.NewTextResponse(
				fmt.Sprintf("references in: %s at line %d, char %d",
					params.FilePath, params.Line, params.Character),
			), nil
		},
	)
}

type referencesParams struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// RestartTool wraps LSP server restart as an eino Tool.
func RestartTool(manager *lsp.Manager) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"lsp_restart",
		"Restart an LSP server by name. If name is empty, restarts all servers.",
		func(ctx context.Context, params restartParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			name := params.Name
			if name == "" {
				name = "all servers"
			}
			// TODO: implement using manager
			return fantasy.NewTextResponse(
				fmt.Sprintf("restarting: %s", strings.Trim(name, "\"")),
			), nil
		},
	)
}

type restartParams struct {
	Name string `json:"name,omitempty"`
}
