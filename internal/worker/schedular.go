package worker

import (
	"database/sql"
	"fmt"

	"time"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
)

func StartScheduler(jobChan chan string, shutdown <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-shutdown:
			for jobID := range jobChan {
				storage.UpdateJobToPending(jobID)
			}
			return

		case <-ticker.C:
			jobID, err := storage.GetPendingJobId()

			if err == nil {
				select {
				case jobChan <- jobID:
					clilogger.LogInfo(fmt.Sprint("dispatched job: ", jobID))
				default:
					storage.UpdateJobToPending(jobID)
				}
			} else if err != sql.ErrNoRows {
				clilogger.LogError("unable to schedule job")
			}
		}
	}
}
