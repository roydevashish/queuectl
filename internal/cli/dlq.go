package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
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
		query := `SELECT id, command, state, attempts, created_at FROM jobs WHERE state = 'dead'`

		rows, _ := storage.DB.Query(query)
		defer rows.Close()

		t := table.New(os.Stdout)
		t.SetHeaders("id", "command", "state", "attempts", "created at")

		for rows.Next() {
			var id, command, state string
			var attempts int
			var created string
			rows.Scan(&id, &command, &state, &attempts, &created)
			t.AddRow(id, command, state, strconv.Itoa(attempts), created)
		}

		t.Render()
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

		result, err := storage.DB.Exec(`
			UPDATE jobs SET state='pending', attempts=0, next_retry_at=NULL, updated_at=datetime('now', '+05 hours', '+30 minutes')
			WHERE id=? AND state='dead'
    `, jobID)
		if err != nil {
			clilogger.LogError(err.Error())
			return
		}

		rowCount, _ := result.RowsAffected()
		if rowCount == 0 {
			clilogger.LogError(fmt.Sprint("invalid job id: ", jobID))
			return
		}

		clilogger.LogInfo(fmt.Sprint("job moved back to pending with job id:", jobID))
	},
}

func init() {
	DLQCmd.AddCommand(listCmd, retryCmd)
}
