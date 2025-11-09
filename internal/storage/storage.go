package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/types"
	"github.com/roydevashish/queuectl/internal/utils"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./queuectl.db?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		clilogger.LogError("unable to open connection to DB")
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		clilogger.LogError("unable to connect to DB")
		log.Fatal(err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			command TEXT NOT NULL,
			state TEXT DEFAULT 'pending',
			attempts INTEGER DEFAULT 0,
			max_retries INTEGER DEFAULT 3,
			base_backoff INTEGER DEFAULT 2,
			next_retry_at TEXT,
			locked_at TEXT,
			output TEXT,
			created_at TEXT DEFAULT (datetime('now', '+05 hours', '+30 minutes')),
			updated_at TEXT DEFAULT (datetime('now', '+05 hours', '+30 minutes'))
		);
			
		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT
		);
				
		INSERT OR IGNORE INTO config (key, value) VALUES ('max_retries', '3');
		INSERT OR IGNORE INTO config (key, value) VALUES ('base_backoff', '2');
	`)
	if err != nil {
		clilogger.LogError("unable to create initial tables")
		log.Fatal(err)
	}
}

func SetConfig(key, value string) error {
	_, err := DB.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, value)

	if err != nil {
		return fmt.Errorf("unable to set configuration")
	}
	return nil
}

func GetJobListFilterByState(state string) ([]types.Job, error) {
	query := `SELECT id, command, state, attempts, created_at FROM jobs`
	if state != "" {
		query += ` WHERE state = '` + state + `'`
	}
	query += ` ORDER BY created_at DESC LIMIT 20`

	rows, _ := DB.Query(query)
	defer rows.Close()

	jobs := make([]types.Job, 0)
	for rows.Next() {
		var job types.Job
		var createdAt string
		err := rows.Scan(&job.ID, &job.Command, &job.State, &job.Attempts, &createdAt)
		if err != nil {
			return nil, err
		}

		job.CreatedAt = utils.ParseTime(createdAt)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func MoveDeadJobToPending(jobID string) error {
	result, err := DB.Exec(`
			UPDATE jobs SET state='pending', attempts=0, next_retry_at=NULL, updated_at=datetime('now', '+05 hours', '+30 minutes')
			WHERE id=? AND state='dead'
    `, jobID)
	if err != nil {
		return err
	}

	rowCount, _ := result.RowsAffected()
	if rowCount == 0 {
		errString := fmt.Sprint("invalid job id: ", jobID)
		return fmt.Errorf(errString)
	}
	return nil
}

func InsertJob(job *types.Job) error {
	_, err := DB.Exec(`
		INSERT INTO jobs (id, command, state, max_retries, base_backoff)
		VALUES (?, ?, 'pending', 
			(SELECT value FROM config WHERE key='max_retries'),
			(SELECT value FROM config WHERE key='base_backoff')
		)
	`, job.ID, job.Command)
	if err != nil {
		if err.Error() != "" && err.Error()[0] == 'U' {
			return fmt.Errorf(fmt.Sprint("job already exists with job id: ", job.ID))
		}

		return fmt.Errorf("unable to enqueue job")
	}
	return nil
}

func GetJobCountByState() map[string]int {
	rows, _ := DB.Query(`SELECT state, COUNT(*) FROM jobs GROUP BY state`)
	defer rows.Close()
	countsMap := map[string]int{}
	for rows.Next() {
		var state string
		var count int
		rows.Scan(&state, &count)
		countsMap[state] = count
	}

	return countsMap
}

func GetJobByJobId(jobID string) (*types.Job, error) {
	var job types.Job
	job.ID = jobID

	err := DB.QueryRow(`
		SELECT command, attempts, max_retries, base_backoff 
		FROM jobs 
		WHERE id=?
  `, job.ID).Scan(&job.Command, &job.Attempts, &job.MaxRetries, &job.BaseBackoff)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprint("job not found with job id: ", jobID))
	}
	return &job, nil
}

func UpdateJobToComplete(jobID, outputStr string) {
	DB.Exec(`
		UPDATE jobs 
		SET state='completed', output=?, locked_at=NULL, updated_at=datetime('now', '+05 hours', '+30 minutes')
		WHERE id=?
  `, outputStr, jobID)
}

func UpdateJobToDead(jobID string, attempts int, outputStr string, runErr error) {
	DB.Exec(`
		UPDATE jobs 
		SET state='dead', attempts=?, output=?, locked_at=NULL
		WHERE id=?
  `, attempts, outputStr+"\nError: "+runErr.Error(), jobID)
}

func UpdateJobToFailed(jobID string, attempts int, next time.Time, outputStr string, runErr error) {
	DB.Exec(`
		UPDATE jobs 
		SET state='failed', attempts=?, next_retry_at=?, output=?, locked_at=NULL
		WHERE id=?
  `, attempts, next.Format("2006-01-02 15:04:05"), outputStr+"\nError: "+runErr.Error(), jobID)
}

func UpdateJobToPending(jobID string) {
	DB.Exec(`UPDATE jobs SET state='pending', locked_at=NULL WHERE id=?`, jobID)
}

func GetPendingJobId() (string, error) {
	var jobID string
	err := DB.QueryRow(`
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
	if err != nil {
		return "", err
	}

	return jobID, nil
}
