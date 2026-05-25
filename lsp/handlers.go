package lsp

import (
	"context"
	"encoding/json"
	"log/slog"

	powernap "github.com/charmbracelet/x/powernap/pkg/lsp"
	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
)

// HandleWorkspaceConfiguration handles workspace/configuration requests.
func HandleWorkspaceConfiguration(_ context.Context, _ string, params json.RawMessage) (any, error) {
	return []map[string]any{{}}, nil
}

// HandleRegisterCapability handles client/registerCapability requests.
func HandleRegisterCapability(_ context.Context, _ string, params json.RawMessage) (any, error) {
	var registerParams protocol.RegistrationParams
	if err := json.Unmarshal(params, &registerParams); err != nil {
		slog.Error("Error unmarshaling registration params", "error", err)
		return nil, err
	}

	for _, reg := range registerParams.Registrations {
		switch reg.Method {
		case "workspace/didChangeWatchedFiles":
			optionsJSON, err := json.Marshal(reg.RegisterOptions)
			if err != nil {
				slog.Error("Error marshaling registration options", "error", err)
				continue
			}
			var options protocol.DidChangeWatchedFilesRegistrationOptions
			if err := json.Unmarshal(optionsJSON, &options); err != nil {
				slog.Error("Error unmarshaling registration options", "error", err)
				continue
			}
			notifyFileWatchRegistration(reg.ID, options.Watchers)
		}
	}
	return nil, nil
}

// FileWatchRegistrationHandler is called when file watch registrations are received.
type FileWatchRegistrationHandler func(id string, watchers []protocol.FileSystemWatcher)

var fileWatchHandler FileWatchRegistrationHandler

// RegisterFileWatchHandler sets the handler for file watch registrations.
func RegisterFileWatchHandler(handler FileWatchRegistrationHandler) {
	fileWatchHandler = handler
}

func notifyFileWatchRegistration(id string, watchers []protocol.FileSystemWatcher) {
	if fileWatchHandler != nil {
		fileWatchHandler(id, watchers)
	}
}

// HandleApplyEdit creates a handler for workspace/applyEdit requests.
func HandleApplyEdit(encoding powernap.OffsetEncoding) func(_ context.Context, _ string, params json.RawMessage) (any, error) {
	return func(_ context.Context, _ string, params json.RawMessage) (any, error) {
		var edit protocol.ApplyWorkspaceEditParams
		if err := json.Unmarshal(params, &edit); err != nil {
			return nil, err
		}
		slog.Debug("workspace/applyEdit requested", "edit", edit.Label, "encoding", encoding)
		return protocol.ApplyWorkspaceEditResult{Applied: false, FailureReason: "not implemented in standalone library"}, nil
	}
}

// HandleServerMessage handles window/showMessage notifications.
func HandleServerMessage(_ context.Context, method string, params json.RawMessage) {
	var msg protocol.ShowMessageParams
	if err := json.Unmarshal(params, &msg); err != nil {
		slog.Debug("Error unmarshal server message", "error", err)
		return
	}

	switch msg.Type {
	case protocol.Error:
		slog.Error("LSP Server", "message", msg.Message)
	case protocol.Warning:
		slog.Warn("LSP Server", "message", msg.Message)
	case protocol.Info:
		slog.Info("LSP Server", "message", msg.Message)
	case protocol.Log:
		slog.Debug("LSP Server", "message", msg.Message)
	}
}

// HandleDiagnostics handles textDocument/publishDiagnostics notifications.
func HandleDiagnostics(client *Client, params json.RawMessage) {
	var diagParams protocol.PublishDiagnosticsParams
	if err := json.Unmarshal(params, &diagParams); err != nil {
		slog.Error("Error unmarshaling diagnostics params", "error", err)
		return
	}

	client.diagnostics.Set(diagParams.URI, diagParams.Diagnostics)

	totalCount := 0
	for _, diagnostics := range client.diagnostics.Seq2() {
		totalCount += len(diagnostics)
	}

	if client.onDiagnosticsChanged != nil {
		client.onDiagnosticsChanged(client.name, totalCount)
	}
}