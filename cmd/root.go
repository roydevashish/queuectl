package cmd

import (
	"os"

	"github.com/roydevashish/queuectl/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "a CLI-based background job queue system",
	Long: `A system that manage background jobs with worker processes, handle
retries using exponential backoff, and maintain a Dead Letter Queue
(DLQ) for permanently failed jobs.`,
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cli.InitCommands(rootCmd)
}
