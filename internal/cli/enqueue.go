package cli

import (
	"fmt"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/roydevashish/queuectl/internal/utils"
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
		rawPayload := args[0]

		job, err := utils.ParseJob(rawPayload)
		if err != nil {
			clilogger.LogError(err.Error())
			return
		}

		if err := storage.InsertJob(job); err != nil {
			clilogger.LogError(err.Error())
			return
		}

		clilogger.LogSuccess(fmt.Sprint("job enqueued with job id: ", job.ID))
	},
}
