package cli

import (
	"encoding/json"
	"fmt"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/roydevashish/queuectl/internal/types"
	"github.com/spf13/cobra"
)

var EnqueueCmd = &cobra.Command{
	Use:   "enqueue [json]",
	Short: "enqueue a new job",
	Long: `The enqueue command serializes a JSON payload representing a job and 
adds it to the backend queue. The payload must contain at least an id
(unique job identifier) and a command (the shell command or script to
execute). 

Jobs are placed in the pending state and become eligible for worker 
to pickup immediately.`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get job
		rawPayload := args[0]

		// unmarshal
		var payload types.EnqueuePayload
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			clilogger.LogError("invalid job details")
			return
		}

		// validate
		if payload.Command == "" {
			clilogger.LogError("job command is required")
			return
		}

		if payload.ID == "" {
			clilogger.LogError("job id is required")
			return
		}

		// store to DB
		_, err := storage.DB.Exec(`
			INSERT INTO jobs (id, command, state, max_retries, base_backoff)
			VALUES (?, ?, 'pending', 
				(SELECT value FROM config WHERE key='max_retries'),
				(SELECT value FROM config WHERE key='base_backoff')
			)
		`, payload.ID, payload.Command)
		if err != nil {
			clilogger.LogError("unable to enqueue job")

			if err.Error() != "" && err.Error()[0] == 'U' {
				clilogger.LogError(fmt.Sprint("job already exists with job id: ", payload.ID))
			}
			return
		}

		clilogger.LogSuccess(fmt.Sprint("job enqueued with job id: ", payload.ID))
	},
}
