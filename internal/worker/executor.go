package worker

import (
	"fmt"
	"os/exec"
	"time"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
)

func ExecuteJob(jobID string) {
	var command string
	var attempts, maxRetries, baseBackoff int

	err := storage.DB.QueryRow(`
		SELECT command, attempts, max_retries, base_backoff 
		FROM jobs 
		WHERE id=?
  `, jobID).Scan(&command, &attempts, &maxRetries, &baseBackoff)
	if err != nil {
		clilogger.LogError(fmt.Sprint("job not found with job id: ", jobID))
		return
	}

	cmd := exec.Command("sh", "-c", command)
	output, runErr := cmd.CombinedOutput()
	outputStr := string(output)

	if runErr == nil {
		storage.DB.Exec(`
			UPDATE jobs 
			SET state='completed', output=?, locked_at=NULL, updated_at=datetime('now', '+05 hours', '+30 minutes')
			WHERE id=?
    `, outputStr, jobID)

		clilogger.LogSuccess(fmt.Sprint("job completed with job id: ", jobID))
	} else {
		attempts++
		clilogger.LogError(fmt.Sprint("job failed with job id: ", jobID))

		if attempts >= maxRetries {
			storage.DB.Exec(`
				UPDATE jobs 
				SET state='dead', attempts=?, output=?, locked_at=NULL
				WHERE id=?
      `, attempts, outputStr+"\nError: "+runErr.Error(), jobID)

			clilogger.LogError(fmt.Sprint("job moved to dlq with job id: ", jobID))
		} else {
			delay := 1
			for i := 0; i < attempts; i++ {
				delay *= baseBackoff
			}

			next := time.Now().Add(time.Duration(delay) * time.Second)
			storage.DB.Exec(`
				UPDATE jobs 
				SET state='failed', attempts=?, next_retry_at=?, output=?, locked_at=NULL
				WHERE id=?
      `, attempts, next.Format("2006-01-02 15:04:05"), outputStr+"\nError: "+runErr.Error(), jobID)

			clilogger.LogAlert(fmt.Sprint("retry job with job id: ", jobID, " in ", delay, " seconds"))
		}
	}
}
