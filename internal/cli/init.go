package cli

import "github.com/spf13/cobra"

func InitCommands(root *cobra.Command) {
	root.AddCommand(EnqueueCmd)
	root.AddCommand(WorkerCmd)
	root.AddCommand(ListCmd)
	root.AddCommand(StatusCmd)
	root.AddCommand(DLQCmd)
	root.AddCommand(ConfigCmd)
}
