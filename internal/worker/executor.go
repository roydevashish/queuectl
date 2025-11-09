package worker

import (
	"fmt"
	"os/exec"
	"time"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
)

func ExecuteJob(jobID string) {
	job, err := storage.GetJobByJobId(jobID)
	if err != nil {
		clilogger.LogError(err.Error())
		return
	}

	cmd := exec.Command("sh", "-c", job.Command)
	output, runErr := cmd.CombinedOutput()
	outputStr := string(output)

	if runErr == nil {
		storage.UpdateJobToComplete(jobID, outputStr)
		clilogger.LogSuccess(fmt.Sprint("job completed with job id: ", jobID))
	} else {
		job.Attempts++
		clilogger.LogError(fmt.Sprint("job failed with job id: ", jobID))

		if job.Attempts >= job.MaxRetries {
			storage.UpdateJobToDead(job.ID, job.Attempts, outputStr, runErr)
			clilogger.LogError(fmt.Sprint("job moved to dlq with job id: ", jobID))
		} else {
			delay := 1
			for i := 0; i < job.Attempts; i++ {
				delay *= job.BaseBackoff
			}
			next := time.Now().Add(time.Duration(delay) * time.Second)

			storage.UpdateJobToFailed(job.ID, job.Attempts, next, outputStr, runErr)
			clilogger.LogAlert(fmt.Sprint("retry job with job id: ", jobID, " in ", delay, " seconds"))
		}
	}
}
