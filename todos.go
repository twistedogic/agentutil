package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/twistedogic/agentutil/tools/todos"
)

func newTodoCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Manage the todo list",
	}
	cmd.PersistentFlags().StringVarP(&file, "file", "f", ".todos.json", "Path to todos state file")

	update := &cobra.Command{
		Use:   "update <json>",
		Short: "Replace the todo list from a JSON array and persist to file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []todos.TodoItem
			if err := json.Unmarshal([]byte(args[0]), &items); err != nil {
				return err
			}
			store := todos.Store{Path: file}
			resp, err := todos.Update(store, items)
			if err != nil {
				return err
			}
			return writeJSON(resp)
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List current todos from the state file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := todos.Store{Path: file}
			items, err := store.Load()
			if err != nil {
				return err
			}
			return writeJSON(items)
		},
	}

	cmd.AddCommand(update, list)
	return cmd
}
