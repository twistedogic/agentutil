package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	powernap "github.com/charmbracelet/x/powernap/pkg/lsp"
	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
	"github.com/charmbracelet/x/powernap/pkg/transport"
)

// DiagnosticCounts holds the count of diagnostics by severity.
type DiagnosticCounts struct {
	Error       int
	Warning     int
	Information int
	Hint        int
}

// Client is an LSP client wrapper.
type Client struct {
	client *powernap.Client
	name   string

	cwd       string
	fileTypes []string
	config    ServerConfig

	ctx      context.Context
	resolver VariableResolver

	onDiagnosticsChanged func(name string, count int)

	diagnostics      *VersionedMap[protocol.DocumentURI, []protocol.Diagnostic]
	diagCountsCache  DiagnosticCounts
	diagCountsVersion uint64
	diagCountsMu     sync.Mutex

	openFiles *Map[string, *OpenFileInfo]

	serverState atomic.Value
}

// OpenFileInfo contains information about an open file.
type OpenFileInfo struct {
	Version int32
	URI     protocol.DocumentURI
}

// New creates a new LSP client.
func New(
	ctx context.Context,
	name string,
	cfg ServerConfig,
	resolver VariableResolver,
	cwd string,
) (*Client, error) {
	client := &Client{
		name:        name,
		fileTypes:   cfg.FileTypes,
		diagnostics: NewVersionedMap[protocol.DocumentURI, []protocol.Diagnostic](),
		openFiles:   NewMap[string, *OpenFileInfo](),
		config:      cfg,
		ctx:         ctx,
		resolver:    resolver,
		cwd:         cwd,
	}
	client.serverState.Store(StateStopped)

	if err := client.createPowernapClient(); err != nil {
		return nil, err
	}

	return client, nil
}

// Initialize initializes the LSP client.
func (c *Client) Initialize(ctx context.Context, workspaceDir string) (*protocol.InitializeResult, error) {
	if err := c.client.Initialize(ctx, false); err != nil {
		return nil, fmt.Errorf("failed to initialize the lsp client: %w", err)
	}

	caps := c.client.GetCapabilities()
	protocolCaps := protocol.ServerCapabilities{
		TextDocumentSync: caps.TextDocumentSync,
		CompletionProvider: func() *protocol.CompletionOptions {
			if caps.CompletionProvider != nil {
				return &protocol.CompletionOptions{
					TriggerCharacters:    caps.CompletionProvider.TriggerCharacters,
					AllCommitCharacters:   caps.CompletionProvider.AllCommitCharacters,
					ResolveProvider:       caps.CompletionProvider.ResolveProvider,
				}
			}
			return nil
		}(),
	}

	result := &protocol.InitializeResult{
		Capabilities: protocolCaps,
	}

	c.registerHandlers()
	return result, nil
}

const closeTimeout = 5 * time.Second

// Close closes all open files and shuts down gracefully.
func (c *Client) Close(ctx context.Context) error {
	c.CloseAllFiles(ctx)

	closeCtx, cancel := context.WithTimeout(ctx, closeTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		if err := c.client.Shutdown(closeCtx); err != nil {
			slog.Warn("Failed to shutdown LSP client", "error", err)
		}
		done <- c.client.Exit()
	}()

	select {
	case err := <-done:
		return err
	case <-closeCtx.Done():
		c.client.Kill()
		return closeCtx.Err()
	}
}

func (c *Client) createPowernapClient() error {
	rootURI := string(protocol.URIFromPath(c.cwd))

	command, err := c.resolver.ResolveValue(c.config.Command)
	if err != nil {
		return fmt.Errorf("invalid lsp command: %w", err)
	}

	envs := make([]string, 0, len(c.config.Environment))
	for k, v := range c.config.Environment {
		envs = append(envs, k+"="+v)
	}

	clientConfig := powernap.ClientConfig{
		Command:     command,
		Args:        c.config.Args,
		RootURI:     rootURI,
		Environment: envs,
		Settings:    c.config.Settings,
		InitOptions: c.config.InitOptions,
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{URI: rootURI, Name: filepath.Base(c.cwd)},
		},
	}

	powernapClient, err := powernap.NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create lsp client: %w", err)
	}

	c.client = powernapClient
	return nil
}

func (c *Client) registerHandlers() {
	c.RegisterServerRequestHandler("workspace/applyEdit", HandleApplyEdit(c.client.GetOffsetEncoding()))
	c.RegisterServerRequestHandler("workspace/configuration", HandleWorkspaceConfiguration)
	c.RegisterServerRequestHandler("client/registerCapability", HandleRegisterCapability)
	c.RegisterNotificationHandler("window/showMessage", func(ctx context.Context, method string, params json.RawMessage) {
		HandleServerMessage(ctx, method, params)
	})
	c.RegisterNotificationHandler("textDocument/publishDiagnostics", func(_ context.Context, _ string, params json.RawMessage) {
		HandleDiagnostics(c, params)
	})
}

// Restart closes and recreates the client with the same config.
func (c *Client) Restart() error {
	var openFiles []string
	for uri := range c.openFiles.Seq2() {
		openFiles = append(openFiles, string(uri))
	}

	closeCtx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	if err := c.Close(closeCtx); err != nil {
		slog.Warn("Error closing client during restart", "name", c.name, "error", err)
	}

	c.SetServerState(StateStopped)
	c.diagCountsCache = DiagnosticCounts{}
	c.diagCountsVersion = 0

	if err := c.createPowernapClient(); err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	c.SetServerState(StateStarting)

	if err := c.client.Initialize(initCtx, false); err != nil {
		c.SetServerState(StateError)
		return fmt.Errorf("failed to initialize lsp client: %w", err)
	}

	c.registerHandlers()

	if err := c.WaitForServerReady(initCtx); err != nil {
		slog.Error("Server failed to become ready after restart", "name", c.name, "error", err)
		c.SetServerState(StateError)
		return err
	}

	for _, uri := range openFiles {
		if err := c.OpenFile(initCtx, uri); err != nil {
			slog.Warn("Failed to reopen file after restart", "file", uri, "error", err)
		}
	}
	return nil
}

// ServerState represents the state of an LSP server.
type ServerState int

const (
	StateUnstarted ServerState = iota
	StateStarting
	StateReady
	StateError
	StateStopped
	StateDisabled
)

// GetServerState returns the current state of the LSP server.
func (c *Client) GetServerState() ServerState {
	if val := c.serverState.Load(); val != nil {
		return val.(ServerState)
	}
	return StateStarting
}

// SetServerState sets the current state of the LSP server.
func (c *Client) SetServerState(state ServerState) {
	c.serverState.Store(state)
}

// GetName returns the name of the LSP client.
func (c *Client) GetName() string {
	return c.name
}

// FileTypes returns the file types this LSP client handles.
func (c *Client) FileTypes() []string {
	return slices.Clone(c.fileTypes)
}

// SetDiagnosticsCallback sets the callback for diagnostic changes.
func (c *Client) SetDiagnosticsCallback(callback func(name string, count int)) {
	c.onDiagnosticsChanged = callback
}

// WaitForServerReady waits for the server to be ready.
func (c *Client) WaitForServerReady(ctx context.Context) error {
	c.SetServerState(StateStarting)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	c.openKeyConfigFiles(ctx)

	for {
		select {
		case <-ctx.Done():
			c.SetServerState(StateError)
			return fmt.Errorf("timeout waiting for LSP server to be ready")
		case <-ticker.C:
			if !c.client.IsRunning() {
				continue
			}
			c.SetServerState(StateReady)
			return nil
		}
	}
}

// HandlesFile checks if this LSP client handles the given file.
func (c *Client) HandlesFile(path string) bool {
	if c == nil {
		return false
	}
	if !hasPrefix(path, c.cwd) {
		return false
	}
	return handlesFiletype(c.name, c.fileTypes, path)
}

// hasPrefix is a simple path prefix check.
func hasPrefix(path, prefix string) bool {
	if len(path) < len(prefix) {
		return false
	}
	return path[:len(prefix)] == prefix || path[:len(prefix)+1] == prefix+"/"
}

// OpenFile opens a file in the LSP server.
func (c *Client) OpenFile(ctx context.Context, filepath string) error {
	if !c.HandlesFile(filepath) {
		return nil
	}

	uri := string(protocol.URIFromPath(filepath))

	if _, exists := c.openFiles.Get(uri); exists {
		return nil
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err = c.client.NotifyDidOpenTextDocument(ctx, uri, string(powernap.DetectLanguage(filepath)), 1, string(content)); err != nil {
		return err
	}

	c.openFiles.Set(uri, &OpenFileInfo{
		Version: 1,
		URI:     protocol.DocumentURI(uri),
	})

	return nil
}

// NotifyChange notifies the server about a file change.
func (c *Client) NotifyChange(ctx context.Context, filepath string) error {
	if c == nil {
		return nil
	}
	uri := string(protocol.URIFromPath(filepath))

	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fileInfo, isOpen := c.openFiles.Get(uri)
	if !isOpen {
		return fmt.Errorf("cannot notify change for unopened file: %s", filepath)
	}

	fileInfo.Version++

	changes := []protocol.TextDocumentContentChangeEvent{
		{
			Value: protocol.TextDocumentContentChangeWholeDocument{
				Text: string(content),
			},
		},
	}

	return c.client.NotifyDidChangeTextDocument(ctx, uri, int(fileInfo.Version), changes)
}

// IsFileOpen checks if a file is currently open.
func (c *Client) IsFileOpen(filepath string) bool {
	uri := string(protocol.URIFromPath(filepath))
	_, exists := c.openFiles.Get(uri)
	return exists
}

// CloseAllFiles closes all currently open files.
func (c *Client) CloseAllFiles(ctx context.Context) {
	for uri := range c.openFiles.Seq2() {
		if err := c.client.NotifyDidCloseTextDocument(ctx, uri); err != nil {
			slog.Warn("Error closing file", "uri", uri, "error", err)
			continue
		}
		c.openFiles.Del(uri)
	}
}

// GetDiagnostics returns all diagnostics for all files.
func (c *Client) GetDiagnostics() map[protocol.DocumentURI][]protocol.Diagnostic {
	if c == nil {
		return nil
	}
	return c.diagnostics.Copy()
}

// GetDiagnosticCounts returns cached diagnostic counts by severity.
func (c *Client) GetDiagnosticCounts() DiagnosticCounts {
	if c == nil {
		return DiagnosticCounts{}
	}
	currentVersion := c.diagnostics.Version()

	c.diagCountsMu.Lock()
	defer c.diagCountsMu.Unlock()

	if currentVersion == c.diagCountsVersion {
		return c.diagCountsCache
	}

	counts := DiagnosticCounts{}
	for _, diags := range c.diagnostics.Seq2() {
		for _, diag := range diags {
			switch diag.Severity {
			case protocol.SeverityError:
				counts.Error++
			case protocol.SeverityWarning:
				counts.Warning++
			case protocol.SeverityInformation:
				counts.Information++
			case protocol.SeverityHint:
				counts.Hint++
			}
		}
	}

	c.diagCountsCache = counts
	c.diagCountsVersion = currentVersion
	return counts
}

// OpenFileOnDemand opens a file only if it's not already open.
func (c *Client) OpenFileOnDemand(ctx context.Context, filepath string) error {
	if c == nil {
		return nil
	}
	if c.IsFileOpen(filepath) {
		return nil
	}
	return c.OpenFile(ctx, filepath)
}

// RegisterNotificationHandler registers a notification handler.
func (c *Client) RegisterNotificationHandler(method string, handler transport.NotificationHandler) {
	c.client.RegisterNotificationHandler(method, handler)
}

// RegisterServerRequestHandler handles server requests.
func (c *Client) RegisterServerRequestHandler(method string, handler transport.Handler) {
	c.client.RegisterHandler(method, handler)
}

func (c *Client) openKeyConfigFiles(ctx context.Context) {
	for _, file := range c.config.RootMarkers {
		file = filepath.Join(c.cwd, file)
		if _, err := os.Stat(file); err == nil {
			if err := c.OpenFile(ctx, file); err != nil {
				slog.Error("Failed to open key config file", "file", file, "error", err)
			}
		}
	}
}

// NotifyWorkspaceChange sends a workspace-level file change notification.
func (c *Client) NotifyWorkspaceChange(ctx context.Context) error {
	if c == nil {
		return nil
	}
	return c.client.NotifyDidChangeWatchedFiles(ctx, []protocol.FileEvent{
		{URI: protocol.DocumentURI(protocol.URIFromPath(c.cwd)), Type: protocol.Changed},
	})
}

// RefreshOpenFiles re-notifies the LSP server about all open files.
func (c *Client) RefreshOpenFiles(ctx context.Context) {
	if c == nil {
		return
	}
	for uri, info := range c.openFiles.Seq2() {
		path, err := protocol.DocumentURI(uri).Path()
		if err != nil {
			slog.Warn("Failed to convert URI to path", "uri", uri, "error", err)
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("Failed to read file for refresh", "path", path, "error", err)
			continue
		}
		info.Version++
		changes := []protocol.TextDocumentContentChangeEvent{
			{
				Value: protocol.TextDocumentContentChangeWholeDocument{
					Text: string(content),
				},
			},
		}
		if err := c.client.NotifyDidChangeTextDocument(ctx, uri, int(info.Version), changes); err != nil {
			slog.Warn("Failed to notify file change", "uri", uri, "error", err)
		}
	}
}

// WaitForDiagnostics waits until diagnostics stop changing.
func (c *Client) WaitForDiagnostics(ctx context.Context, timeout time.Duration) {
	if c == nil {
		return
	}

	const (
		firstChangeDuration = 1 * time.Second
		settleDuration      = 300 * time.Millisecond
	)

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	firstChangeTimer := time.NewTimer(min(timeout, firstChangeDuration))
	defer firstChangeTimer.Stop()
	previousVersion := c.diagnostics.Version()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-deadline.C:
			return
		case <-firstChangeTimer.C:
			return
		case <-ticker.C:
			currentVersion := c.diagnostics.Version()
			if currentVersion != previousVersion {
				c.waitForDiagnosticsToSettle(ctx, deadline.C, settleDuration)
				return
			}
		}
	}
}

func (c *Client) waitForDiagnosticsToSettle(ctx context.Context, deadline <-chan time.Time, settleDuration time.Duration) {
	lastVersion := c.diagnostics.Version()
	settleTicker := time.NewTicker(50 * time.Millisecond)
	defer settleTicker.Stop()
	stableStart := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-deadline:
			return
		case <-settleTicker.C:
			currentVersion := c.diagnostics.Version()
			if currentVersion != lastVersion {
				lastVersion = currentVersion
				stableStart = time.Now()
			} else if time.Since(stableStart) >= settleDuration {
				return
			}
		}
	}
}

// FindReferences finds all references to the symbol at the given position.
func (c *Client) FindReferences(ctx context.Context, filepath string, line, character int, includeDeclaration bool) ([]protocol.Location, error) {
	if err := c.OpenFileOnDemand(ctx, filepath); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.FindReferences(ctx, filepath, line-1, character-1, includeDeclaration)
}