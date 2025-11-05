package cli

import (
	"fmt"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/spf13/cobra"
)

var DLQCmd = &cobra.Command{
	Use:   "dlq",
	Short: "manage the Dead-Letter Queue (DLQ) for jobs that failed after exhausting all retries",
	Long: `The dlq subcommand provides tools to list and retry jobs that have
exceeded the maximum retry attempts. Jobs in the DLQ are no longer
automatically retried but retain full payload, error history, etc. 

Use list to view failed jobs and retry to requeue them for processing.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists jobs that landed in the dead-letter queue",
	Long: `Shows jobs that exceeded max retries or were manually moved to DLQ.
Includes failure reason, last error message, and full job payload.`,

	Run: func(cmd *cobra.Command, args []string) {
		clilogger.LogSuccess("show dlq list")
	},
}

var retryCmd = &cobra.Command{
	Use:   "retry [job_id]",
	Short: "retries a specific DLQ job",
	Long: `Moves the job back to the pending state with reset attempt counter
(or incremented, based on config). Optionally clears error history.`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		clilogger.LogAlert(fmt.Sprint("retry dlq job: ", jobID))
	},
}

func init() {
	DLQCmd.AddCommand(listCmd, retryCmd)
}
