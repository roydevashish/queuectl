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
				storage.DB.Exec(`UPDATE jobs SET state='pending', locked_at=NULL WHERE id=?`, jobID)
			}

			return
		case <-ticker.C:
			var jobID string
			err := storage.DB.QueryRow(`
				UPDATE jobs SET 
					state='processing',
					locked_at=datetime('now', '+05 hours', '+30 minutes'),
					updated_at=datetime('now', '+05 hours', '+30 minutes')
				WHERE id = (
					SELECT id FROM jobs 
					WHERE (state='pending' OR state='failed')
						AND (next_retry_at IS NULL OR next_retry_at <= datetime('now', '+05 hours', '+30 minutes'))
					ORDER BY created_at ASC LIMIT 1
				)
				RETURNING id
      `).Scan(&jobID)

			if err == nil {
				select {
				case jobChan <- jobID:
					clilogger.LogInfo(fmt.Sprint("dispatched job: ", jobID))
				default:
					storage.DB.Exec(`UPDATE jobs SET state='pending', locked_at=NULL WHERE id=?`, jobID)
				}
			} else if err != sql.ErrNoRows {
				clilogger.LogError("unable to schedule job")
			}
		}
	}
}
