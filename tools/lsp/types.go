package lsp

// ResolvedConfig holds the LSP configuration resolved from a ConfigStore.
// Used by Goof-based applications that need full resolution.
type ResolvedConfig struct {
	Servers map[string]ResolvedServerConfig
	AutoLSP bool
}

// ResolvedServerConfig is a resolved LSP server configuration ready for client creation.
type ResolvedServerConfig struct {
	Command     string
	Args        []string
	Environment map[string]string
	FileTypes   []string
	RootMarkers []string
	InitOptions map[string]any
	Settings    map[string]any
	Timeout     int
	FileWatchers []FileWatcher
}

// FileWatcher describes a file pattern to watch.
type FileWatcher struct {
	Kind float64
	Glob string
}

// NewResolvedConfig creates a default resolved config with common LSP servers.
func NewResolvedConfig() *ResolvedConfig {
	return &ResolvedConfig{
		Servers: map[string]ResolvedServerConfig{
			"gopls": {
				Command:   "gopls",
				FileTypes: []string{".go", ".mod", ".sum"},
				RootMarkers: []string{"go.mod"},
			},
			"typescript-language-server": {
				Command:   "typescript-language-server",
				Args:      []string{"--typescript-preferences.useLibraryUsagesForExtractedFiles=true"},
				FileTypes: []string{".ts", ".tsx", ".js", ".jsx", ".json"},
				RootMarkers: []string{"package.json", "tsconfig.json"},
			},
			"rust-analyzer": {
				Command:   "rust-analyzer",
				FileTypes: []string{".rs", ".toml"},
				RootMarkers: []string{"Cargo.toml"},
			},
			"clangd": {
				Command:   "clangd",
				FileTypes: []string{".c", ".cpp", ".h", ".hpp", ".cc"},
				RootMarkers: []string{".clangd", "compile_commands.json"},
			},
			"jedi-language-server": {
				Command:   "jedi-language-server",
				FileTypes: []string{".py"},
				RootMarkers: []string{"setup.py", "pyproject.toml", "requirements.txt"},
			},
		},
		AutoLSP: true,
	}
}

// AddServer adds or overwrites a server configuration.
func (c *ResolvedConfig) AddServer(name string, cfg ResolvedServerConfig) {
	if c.Servers == nil {
		c.Servers = make(map[string]ResolvedServerConfig)
	}
	c.Servers[name] = cfg
}

// IsEmpty returns true if no servers are configured.
func (c *ResolvedConfig) IsEmpty() bool {
	return len(c.Servers) == 0
}