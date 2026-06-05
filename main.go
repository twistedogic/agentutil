package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "agentutil",
		Short:         "Agent utility CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newLSPCmd())
	root.AddCommand(newFetchCmd())
	root.AddCommand(newWikiCmd())
	root.AddCommand(newSearchCmd())
	root.AddCommand(newSkillCmd())
	root.AddCommand(newTodoCmd())

	if err := root.Execute(); err != nil {
		writeError(err)
		os.Exit(1)
	}
}
