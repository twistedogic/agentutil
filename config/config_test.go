package config

import "testing"

// TestServerConfigDefaults tests ServerConfig field defaults.
func TestServerConfigDefaults(t *testing.T) {
	cfg := ServerConfig{
		Command: "gopls",
	}
	if cfg.Disabled {
		t.Error("Disabled should be false by default")
	}
	if cfg.Timeout != 0 {
		t.Error("Timeout should be 0 by default")
	}
}

// TestConfigMap tests the Config map type.
func TestConfigMap(t *testing.T) {
	cfg := Config{
		"gopls": {
			Command: "gopls",
			Args:    []string{"-v"},
		},
		"typescript-language-server": {
			Command: "typescript-language-server",
			Args:    []string{"--stdio"},
		},
	}

	if len(cfg) != 2 {
		t.Errorf("expected 2 entries, got %d", len(cfg))
	}

	gopls, ok := cfg["gopls"]
	if !ok {
		t.Error("gopls not found")
	}
	if gopls.Command != "gopls" {
		t.Errorf("expected gopls, got %s", gopls.Command)
	}
}
