// Package lsp provides a Manager for Language Server Protocol (LSP) clients.
// Extracted from github.com/charmbracelet/crush for reuse.
package lsp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	powernapconfig "github.com/charmbracelet/x/powernap/pkg/config"
	powernap "github.com/charmbracelet/x/powernap/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// unavailableRetryDelay is the time to wait before retrying an unavailable LSP server.
const unavailableRetryDelay = 30 * time.Second

// skipAutoStartCommands contains commands that are too generic or ambiguous to
// auto-start without explicit user configuration.
var skipAutoStartCommands = map[string]bool{
	"buck2":    true,
	"buf":      true,
	"cue":      true,
	"dart":     true,
	"deno":     true,
	"dotnet":   true,
	"dprint":   true,
	"gleam":    true,
	"java":     true,
	"julia":    true,
	"koka":     true,
	"node":     true,
	"npx":      true,
	"perl":     true,
	"plz":      true,
	"python":   true,
	"python3":  true,
	"R":        true,
	"racket":   true,
	"rome":     true,
	"rubocop":  true,
	"ruff":     true,
	"scarb":    true,
	"solc":     true,
	"stylua":   true,
	"swipl":    true,
	"tflint":   true,
}

// Manager handles lazy initialization of LSP clients based on file types.
type Manager struct {
	clients     *Map[string, *Client]
	unavailable *Map[string, time.Time]
	config      ConfigStore
	manager     *powernapconfig.Manager
	callback    func(name string, client *Client)
	now         func() time.Time
	workDir     string
}

// ConfigStore is the interface for reading LSP configuration.
type ConfigStore interface {
	LSP() map[string]ServerConfig
	AutoLSP() *bool
	Resolver() VariableResolver
}

// VariableResolver resolves configuration variables.
type VariableResolver interface {
	ResolveValue(v string) (string, error)
}

// ServerConfig describes an LSP server configuration.
type ServerConfig struct {
	Command     string
	Args        []string
	Environment map[string]string
	FileTypes   []string
	RootMarkers []string
	InitOptions map[string]any
	Settings    map[string]any
	Timeout     int
}

// NewManager creates a new LSP manager service from a ConfigStore.
func NewManager(config ConfigStore, workDir string) *Manager {
	pm := powernapconfig.NewManager()
	pm.LoadDefaults()

	// Merge user-configured LSPs into the manager.
	for name, serverConfig := range config.LSP() {
		if serverConfig.Disabled {
			slog.Debug("LSP disabled by user config", "name", name)
			pm.RemoveServer(name)
			continue
		}

		// HACK: the user might have the command name in their config instead
		// of the actual name. Find and use the correct name.
		actualName := resolveServerName(pm, name)
		pm.AddServer(actualName, &powernapconfig.ServerConfig{
			Command:     serverConfig.Command,
			Args:        serverConfig.Args,
			Environment: serverConfig.Environment,
			FileTypes:   serverConfig.FileTypes,
			RootMarkers: serverConfig.RootMarkers,
			InitOptions: serverConfig.InitOptions,
			Settings:    serverConfig.Settings,
		})
	}

	return &Manager{
		clients:     NewMap[string, *Client](),
		unavailable: NewMap[string, time.Time](),
		config:      config,
		manager:     pm,
		callback:    func(string, *Client) {},
		now:         time.Now,
		workDir:     workDir,
	}
}

// Clients returns the map of LSP clients.
func (s *Manager) Clients() *Map[string, *Client] {
	return s.clients
}

// SetCallback sets a callback that is invoked when a new LSP client is started.
func (s *Manager) SetCallback(cb func(name string, client *Client)) {
	s.callback = cb
}

// TrackConfigured will callback for user-configured LSPs without creating clients.
func (s *Manager) TrackConfigured() {
	var wg sync.WaitGroup
	for name := range s.manager.GetServers() {
		if !s.isUserConfigured(name) {
			continue
		}
		wg.Go(func() {
			s.callback(name, nil)
		})
	}
	wg.Wait()
}

// Start starts an LSP server that can handle the given file path.
func (s *Manager) Start(ctx context.Context, path string) {
	var wg sync.WaitGroup
	for name, server := range s.manager.GetServers() {
		wg.Go(func() {
			s.startServer(ctx, name, path, server)
		})
	}
	wg.Wait()
}

func (s *Manager) startServer(ctx context.Context, name, path string, server *powernapconfig.ServerConfig) {
	var (
		isUserConfigured = s.isUserConfigured(name)
		autoLSP          = s.config.AutoLSP()
	)
	if !isUserConfigured && autoLSP != nil && !*autoLSP {
		slog.Debug("Auto-start LSP disabled", "name", name)
		return
	}

	cfg := s.buildConfig(name, server)
	if cfg.Disabled {
		return
	}

	if client, ok := s.clients.Get(name); ok {
		switch client.GetServerState() {
		case StateReady, StateStarting, StateDisabled:
			s.callback(name, client)
			return
		}
	}

	if !isUserConfigured {
		if s.recentlyUnavailable(name) {
			return
		}
		if _, err := exec.LookPath(server.Command); err != nil {
			slog.Debug("LSP server not installed, skipping", "name", name, "command", server.Command)
			s.markUnavailable(name)
			return
		}
		s.clearUnavailable(name)
		if skipAutoStartCommands[server.Command] {
			slog.Debug("LSP command too generic for auto-start, skipping", "name", name, "command", server.Command)
			return
		}
	}

	if !handles(server, path, s.workDir) {
		return
	}

	if client, ok := s.clients.Get(name); ok {
		switch client.GetServerState() {
		case StateReady, StateStarting, StateDisabled:
			s.callback(name, client)
			return
		}
	}

	client, err := New(
		ctx,
		name,
		cfg,
		s.config.Resolver(),
		s.workDir,
	)
	if err != nil {
		slog.Error("Failed to create LSP client", "name", name, "error", err)
		return
	}

	if existing, ok := s.clients.Get(name); ok {
		switch existing.GetServerState() {
		case StateReady, StateStarting, StateDisabled:
			_ = client.Close(ctx)
			s.callback(name, existing)
			return
		}
	}
	s.clients.Set(name, client)

	client.serverState.Store(StateStarting)

	initCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	if _, err := client.Initialize(initCtx, s.workDir); err != nil {
		slog.Error("LSP client initialization failed", "name", name, "error", err)
		_ = client.Close(ctx)
		s.clients.Del(name)
		return
	}

	if err := client.WaitForServerReady(initCtx); err != nil {
		slog.Warn("LSP server not fully ready, continuing anyway", "name", name, "error", err)
		client.SetServerState(StateError)
	} else {
		client.SetServerState(StateReady)
	}

	slog.Debug("LSP client started", "name", name)
	s.callback(name, client)
}

func (s *Manager) isUserConfigured(name string) bool {
	cfg, ok := s.config.LSP()[name]
	return ok && !cfg.Disabled
}

func (s *Manager) recentlyUnavailable(name string) bool {
	lastUnavailableAt, exists := s.unavailable.Get(name)
	if !exists {
		return false
	}
	if s.now().Sub(lastUnavailableAt) < unavailableRetryDelay {
		return true
	}
	s.unavailable.Del(name)
	return false
}

func (s *Manager) markUnavailable(name string) {
	s.unavailable.Set(name, s.now())
}

func (s *Manager) clearUnavailable(name string) {
	s.unavailable.Del(name)
}

func (s *Manager) buildConfig(name string, server *powernapconfig.ServerConfig) ServerConfig {
	cfg := ServerConfig{
		Command:     server.Command,
		Args:        server.Args,
		Environment: server.Environment,
		FileTypes:   server.FileTypes,
		RootMarkers: server.RootMarkers,
		InitOptions: server.InitOptions,
		Settings:    server.Settings,
	}
	if userCfg, ok := s.config.LSP()[name]; ok {
		cfg.Timeout = userCfg.Timeout
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30
	}
	return cfg
}

func resolveServerName(manager *powernapconfig.Manager, name string) string {
	if _, ok := manager.GetServer(name); ok {
		return name
	}
	for sname, server := range manager.GetServers() {
		if server.Command == name {
			return sname
		}
	}
	return name
}

func handlesFiletype(sname string, fileTypes []string, filePath string) bool {
	if len(fileTypes) == 0 {
		return true
	}

	kind := powernap.DetectLanguage(filePath)
	name := strings.ToLower(filepath.Base(filePath))
	for _, filetype := range fileTypes {
		suffix := strings.ToLower(filetype)
		if !strings.HasPrefix(suffix, ".") {
			suffix = "." + suffix
		}
		if strings.HasSuffix(name, suffix) || filetype == string(kind) {
			slog.Debug("Handles file", "name", sname, "file", name, "filetype", filetype, "kind", kind)
			return true
		}
	}
	return false
}

func hasRootMarkers(dir string, markers []string) bool {
	if len(markers) == 0 {
		return true
	}
	for _, pattern := range markers {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

func handles(server *powernapconfig.ServerConfig, filePath, workDir string) bool {
	return handlesFiletype(server.Command, server.FileTypes, filePath) &&
		hasRootMarkers(workDir, server.RootMarkers)
}

// KillAll force-kills all LSP clients.
func (s *Manager) KillAll(ctx context.Context) {
	var wg sync.WaitGroup
	for name, client := range s.clients.Seq2() {
		wg.Go(func() {
			defer func() { s.callback(name, client) }()
			client.client.Kill()
			client.SetServerState(StateStopped)
			s.clients.Del(name)
			slog.Debug("Killed LSP client", "name", name)
		})
	}
	wg.Wait()
}

// StopAll gracefully stops all LSP clients.
func (s *Manager) StopAll(ctx context.Context) {
	var wg sync.WaitGroup
	for name, client := range s.clients.Seq2() {
		wg.Go(func() {
			defer func() { s.callback(name, client) }()
			if err := client.Close(ctx); err != nil &&
				!errors.Is(err, io.EOF) &&
				!errors.Is(err, context.Canceled) &&
				!errors.Is(err, jsonrpc2.ErrClosed) &&
				err.Error() != "signal: killed" {
				slog.Warn("Failed to stop LSP client", "name", name, "error", err)
			}
			client.SetServerState(StateStopped)
			s.clients.Del(name)
			slog.Debug("Stopped LSP client", "name", name)
		})
	}
	wg.Wait()
}

// Map is a generic thread-safe map.
type Map[K comparable, V any] struct {
	m    map[K]V
	lock sync.RWMutex
}

// NewMap creates a new Map.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}
}

// Get returns the value for key.
func (m *Map[K, V]) Get(key K) (V, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

// Set sets the value for key.
func (m *Map[K, V]) Set(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[key] = value
}

// Del deletes the key.
func (m *Map[K, V]) Del(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, key)
}

// Len returns the number of entries.
func (m *Map[K, V]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.m)
}

// Seq returns a channel that yields key-value pairs.
func (m *Map[K, V]) Seq() <-chan K {
	ch := make(chan K)
	go func() {
		m.lock.RLock()
		defer m.lock.RUnlock()
		defer close(ch)
		for k := range m.m {
			ch <- k
		}
	}()
	return ch
}

// Seq2 returns all key-value pairs as a map.
func (m *Map[K, V]) Seq2() map[K]V {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := make(map[K]V, len(m.m))
	for k, v := range m.m {
		result[k] = v
	}
	return result
}