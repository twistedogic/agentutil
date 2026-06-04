package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

//go:embed skills/*
var skillsFS embed.FS

func newSkillCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Install agentutil skills",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			path, err = expandPath(path)
			if err != nil {
				return err
			}
			return fs.WalkDir(skillsFS, "skills", func(p string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}
				rel, _ := strings.CutPrefix(filepath.Dir(p), "skills")
				loc := filepath.Join(path, rel)
				if err := os.MkdirAll(loc, 0755); err != nil {
					return err
				}
				b, err := skillsFS.ReadFile(p)
				if err != nil {
					return err
				}
				dst := filepath.Join(loc, filepath.Base(p))
				if err := os.WriteFile(dst, b, 0755); err != nil {
					return err
				}
				slog.Info("created " + dst)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&path, "path", "p", "~/.agents/skills", "Path to install skills. Create if not exist.")
	return cmd

}
