# queuectl - CLI-Based Background Job Queue System

A minimal, production-grade job queue system written in Go. Supports enqueuing jobs, multiple parallel workers, exponential backoff retries, Dead Letter Queue (DLQ), and full persistence across restarts—all controlled via a clean CLI.

## Features
- Persistent SQLite storage (WAL mode for concurrency safety)
- One scheduler goroutine + fixed-size worker pool (backpressure via buffered channel)
- Atomic job locking to prevent duplicate execution
- Exponential backoff retries (`delay = base^attempts seconds`)
- Dead Letter Queue for exhausted retries
- Global configuration (max_retries, base_backoff)
- Graceful shutdown (workers finish current job)
- Comprehensive CLI with help texts

## Setup Instructions

### Prerequisites
- Go 1.24+  
- git

### Install & Build
```bash
# Clone the repo
git clone https://github.com/roydevashish/queuectl.git
cd queuectl

# Tidy dependencies
go mod tidy

# Build the binary
go build -o queuectl

# Make it executable (optional)
chmod +x queuectl
```

### Run Locally
```bash
# Start with the built binary
./queuectl --help
```

The database file `queuectl.db` will be created automatically in the current directory on first run.

## Usage Examples

### 1. Enqueue Jobs
```bash
./queuectl enqueue '{"id":"job0", "command":"echo Hello World"}'
# Output: 2025/11/08 19:36:16 INFO 	✅ job enqueued with job id: job0

./queuectl enqueue '{"id":"failjob","command":"false"}'
# Output: 2025/11/08 19:37:06 INFO 	✅ job enqueued with job id: failjob
```

### 2. Start Workers
```bash
./queuectl worker start --count 5
# Output:
# 2025/11/08 19:39:15 INFO  ℹ️ starting total #2 workers, press Ctrl+C to stop
# 2025/11/08 19:39:15 INFO  ℹ️ worker started with id: 1
# 2025/11/08 19:39:15 INFO  ℹ️ worker started with id: 2
# 2025/11/08 19:39:15 INFO  ℹ️ dispatched job: job0
# 2025/11/08 19:39:15 INFO  ℹ️ dispatched job: failjob
# 2025/11/08 19:39:15 ERROR 🚫 job failed with job id: failjob
# 2025/11/08 19:39:15 INFO  ✅ job completed with job id: job0
# 2025/11/08 19:39:15 INFO  ⚠️ retry job with job id: failjob in 2 seconds
# 2025/11/08 19:39:17 INFO  ℹ️ dispatched job: failjob
# 2025/11/08 19:39:17 ERROR 🚫 job failed with job id: failjob
# 2025/11/08 19:39:17 INFO  ⚠️ retry job with job id: failjob in 4 seconds
# 2025/11/08 19:39:21 INFO  ℹ️ dispatched job: failjob
# 2025/11/08 19:39:21 ERROR 🚫 job failed with job id: failjob
# 2025/11/08 19:39:21 ERROR 🚫 job moved to dlq with job id: failjob
```

### 3. Check Status
```bash
./queuectl status
# Output: while 2 workers running
# 2025/11/08 19:40:58 INFO 	ℹ️ status
|------------|------------|----------------|---------------|------------|---------|
| 💻 workers | ⏳ pending | 🔄 processing  | ✅ completed  | ❌ failed  | 💀 dead |
|------------|------------|----------------|---------------|------------|---------|
|     2      |     0      |       0        |      1        |     0      |    1    |
|------------|------------|----------------|---------------|------------|---------|

# Output: while workers not running
# 2025/11/08 19:43:51 INFO 	ℹ️ status
|------------|------------|----------------|---------------|------------|---------|
| 💻 workers | ⏳ pending | 🔄 processing  | ✅ completed  | ❌ failed  | 💀 dead |
|------------|------------|----------------|---------------|------------|---------|
|     0      |     0      |       0        |      1        |     0      |    1    |
|------------|------------|----------------|---------------|------------|---------|
```

### 4. List Jobs
```bash
./queuectl list --state pending
# Output:
┌───────┬──────────────────┬─────────┬──────────┬─────────────────────┐
│  id   │     command      │  state  │ attempts │     created at      │
├───────┼──────────────────┼─────────┼──────────┼─────────────────────┤
│ job03 │ echo Hello World │ pending │ 0        │ 2025-11-08 19:45:08 │
├───────┼──────────────────┼─────────┼──────────┼─────────────────────┤
│ job02 │ echo Hello World │ pending │ 0        │ 2025-11-08 19:44:58 │
└───────┴──────────────────┴─────────┴──────────┴─────────────────────┘

./queuectl list --state dead
# Output: 
┌─────────┬─────────┬───────┬──────────┬─────────────────────┐
│   id    │ command │ state │ attempts │     created at      │
├─────────┼─────────┼───────┼──────────┼─────────────────────┤
│ failjob │ false   │ dead  │ 3        │ 2025-11-08 19:39:08 │
└─────────┴─────────┴───────┴──────────┴─────────────────────┘
```

### 5. DLQ Management
```bash
./queuectl dlq list
# Output
┌─────────┬─────────┬───────┬──────────┬─────────────────────┐
│   id    │ command │ state │ attempts │     created at      │
├─────────┼─────────┼───────┼──────────┼─────────────────────┤
│ failjob │ false   │ dead  │ 3        │ 2025-11-08 19:39:08 │
└─────────┴─────────┴───────┴──────────┴─────────────────────┘

./queuectl dlq retry failjob
# Output: Job failjob moved back to pending
# 2025/11/08 19:45:51 INFO 	ℹ️ job moved back to pending with job id:bad
```

### 6. Configuration
```bash
./queuectl config set max_retries 5
# Output
# 2025/11/08 19:48:29 INFO 	✅ set config max_retries:5

./queuectl config set base_backoff 3
# New jobs will use these defaults
# 2025/11/08 19:48:49 INFO 	✅ set config base_backoff:3
```

### 7. Full Demo Flow
```bash
./queuectl enqueue '{"id":"job1", "command":"echo Success"}'
./queuectl enqueue '{"id":"job2", "command":"sleep 10"}'
./queuectl enqueue '{"id":"job3", "command":"echo Success"}'
./queuectl enqueue '{"id":"bad","command":"exit 1"}'
./queuectl worker start --count 3

# Output
# 2025/11/08 19:53:20 INFO   ✅ job enqueued with job id: job1
# 2025/11/08 19:53:30 INFO   ✅ job enqueued with job id: job2
# 2025/11/08 19:53:42 INFO   ✅ job enqueued with job id: job3
# 2025/11/08 19:53:56 INFO   ✅ job enqueued with job id: bad
# 2025/11/08 19:54:04 INFO   ℹ️ starting total #3 workers, press Ctrl+C to stop
# 2025/11/08 19:54:04 INFO   ℹ️ worker started with id: 3
# 2025/11/08 19:54:04 INFO   ℹ️ worker started with id: 1
# 2025/11/08 19:54:04 INFO   ℹ️ worker started with id: 2
# 2025/11/08 19:54:04 INFO   ℹ️ dispatched job: job1
# 2025/11/08 19:54:04 INFO   ℹ️ dispatched job: job2
# 2025/11/08 19:54:04 INFO   ℹ️ dispatched job: job3
# 2025/11/08 19:54:04 INFO   ℹ️ dispatched job: bad
# 2025/11/08 19:54:04 INFO   ✅ job completed with job id: job1
# 2025/11/08 19:54:04 INFO   ✅ job completed with job id: job3
# 2025/11/08 19:54:04 ERROR  🚫 job failed with job id: bad
# 2025/11/08 19:54:04 INFO   ⚠️ retry job with job id: bad in 2 seconds
# 2025/11/08 19:54:06 INFO   ℹ️ dispatched job: bad
# 2025/11/08 19:54:06 ERROR  🚫 job failed with job id: bad
# 2025/11/08 19:54:06 INFO   ⚠️ retry job with job id: bad in 4 seconds
# 2025/11/08 19:54:10 INFO   ℹ️ dispatched job: bad
# 2025/11/08 19:54:10 ERROR  🚫 job failed with job id: bad
# 2025/11/08 19:54:10 ERROR  🚫 job moved to dlq with job id: bad
# 2025/11/08 19:54:14 INFO   ✅ job completed with job id: job2
# 2025/11/08 19:54:32 INFO   ℹ️ workers shutting down
# 2025/11/08 19:54:32 INFO   ✅ all workers shutdown
```

```bash
# On another terminal
./queuectl status
# Output:
# 2025/11/08 19:54:20 INFO 	ℹ️ status
┌────────────┬────────────┬───────────────┬──────────────┬───────────┬─────────┐
│ 💻 workers │ ⏳ pending │ 🔄 processing │ ✅ completed │ ❌ failed │ 💀 dead │
├────────────┼────────────┼───────────────┼──────────────┼───────────┼─────────┤
│     3      │     0      │       0       │      3       │     0     │    1    │
└────────────┴────────────┴───────────────┴──────────────┴───────────┴─────────┘

./queuectl dlq list
# Output:
┌─────┬─────────┬───────┬──────────┬─────────────────────┐
│ id  │ command │ state │ attempts │     created at      │
├─────┼─────────┼───────┼──────────┼─────────────────────┤
│ bad │ exit 1  │ dead  │ 3        │ 2025-11-08 19:53:56 │
└─────┴─────────┴───────┴──────────┴─────────────────────┘

./queuectl worker stop
# Output:
# 2025/11/08 19:54:32 INFO 	✅ stoping all workers
```

## Project Structure
```text
|-- cmd
|   |-- root.go
|
|-- internal
|   |-- cli
|   |   |-- config.go
|   |   |-- dlq.go
|   |   |-- enqueue.go
|   |   |-- init.go
|   |   |-- list.go
|   |   |-- status.go
|   |   |-- worker.go
|   |
|   |-- cli_logger
|   |   |-- cli_logger.go
|   |
|   |-- storage
|   |   |-- storage.go
|   |
|   |-- types
|   |   |-- enqueue_payload.go
|   |
|   |-- utils
|   |   |-- utils.go
|   |
|   |-- worker
|       |-- executor.go
|       |-- schedular.go
|
|-- go.mod
|-- go.sum
|-- main.go
|-- README.md
|-- queuectl.db
```

## Architecture Overview

```
                     +------------------+
                     |   CLI (cobra)    |
                     +--------+---------+
                              |
                              v
                    +-------------------+
                    |   SQLite DB       |<-------------------+
                    | (queuectl.db)     |                    |
                    +--------+----------+                    |
                             ^                               |
                             |                               |
             +---------------+---------------+               |
             |                               |               |
             v                               v               |
   +-----------------+             +-----------------+       |
   |  Scheduler (1   |             |  Worker Pool    |-------+
   |  goroutine)     |             |  (N goroutines) |
   +---------+-------+             +---------+-------+
             |                                 |
             |  fixed buffered channel         |
             +-------------------------------> +
```

### Job Lifecycle
```
pending → (scheduler picks & locks) → processing → worker executes
   ↓                                          ↑
   │                                          │
   └─> completed (success)                    └─> failed
                                                  ↓
                                                  retry (exponential backoff)
                                                  ↓
                                              max_retries exceeded → dead (DLQ)
```

- **Scheduler**: Runs every 1s, atomically selects & locks the oldest ready job (`UPDATE ... RETURNING`).
- **Workers**: Read from fixed channel (backpressure). Execute `sh -c <command>`. Update job atomically.
- **Persistence**: All state in SQLite (WAL + synchronous=NORMAL). Survives crashes/restarts.
- **Locking**: `state=processing` + `locked_at` prevents duplicates even across process restarts.
- **Graceful Shutdown**: Close shutdown chan → scheduler stops → workers drain channel → finish current job.

## Assumptions & Trade-offs

### Assumptions
- Jobs are shell commands (`sh -c ...`). Sufficient for most background tasks.
- Single-node deployment (no distributed workers).
- SQLite is acceptable for persistence (excellent for single machine, low-to-medium throughput).
- Jobs are idempotent or user handles non-idempotency.

### Decisions & Trade-offs
- **SQLite over JSON file**: Proper transactions, concurrency safety, and easy querying. WAL mode allows concurrent reads/writes.
- **One scheduler + buffered channel**: Simpler than full work-stealing; backpressure prevents DB overload.
- **No separate "failed" state**: Retries stay in `pending` with `next_retry_at`. Simplifies queries.
- **Exponential backoff = base^attempts**: Pure exponential (2^3 = 8s, etc.). Could add jitter but kept minimal.
- **Global config in DB**: Easy to change without restart; new jobs pick up latest values.
- **No job timeout**: Omitted for simplicity.
- **No priority/scheduled jobs**: Omitted for simplicity.

### Simplifications
- DLQ `list` reuses `list --state dead`.
- Config only supports `max_retries` and `base_backoff`.
- No pagination on list (LIMIT 20).
