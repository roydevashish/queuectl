package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
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
