// Package config provides LSP configuration types.
package config

// ServerConfig describes an LSP server configuration.
type ServerConfig struct {
	Disabled    bool              `json:"disabled,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Environment map[string]string `json:"env,omitempty"`
	FileTypes   []string          `json:"filetypes,omitempty"`
	RootMarkers []string          `json:"root_markers,omitempty"`
	InitOptions map[string]any    `json:"init_options,omitempty"`
	Options     map[string]any    `json:"options,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
}

// Config holds all LSP server configurations.
type Config map[string]ServerConfig