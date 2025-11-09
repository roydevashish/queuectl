package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/aquasecurity/table"
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/types"
)

func WriteFile(file string, data string) error {
	return os.WriteFile(file, []byte(data), 0644)
}

func ReadFile(file string) (int, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(b))
}

func IsRunning(pidFile string) (bool, int) {
	pid, err := ReadFile(pidFile)
	if err != nil {
		return false, 0
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}
	if err = proc.Signal(syscall.Signal(0)); err != nil {
		return false, 0
	}
	return true, pid
}

func ActiveWorkers() int {
	running, _ := IsRunning("pid")
	if !running {
		return 0
	}

	workerCount, err := ReadFile("worker")
	if err != nil {
		return 0
	}

	return workerCount
}

func ParseJob(payload string) (*types.Job, error) {
	var job types.Job
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		return nil, fmt.Errorf("invalid job details")
	}

	if job.Command == "" {
		return nil, fmt.Errorf("job command is required")
	}

	if job.ID == "" {
		return nil, fmt.Errorf("job id is required")
	}

	return &job, nil
}

func PrintJobs(jobs []types.Job) {
	t := table.New(os.Stdout)
	t.SetHeaders("id", "command", "state", "attempts", "created at")
	for _, job := range jobs {
		t.AddRow(job.ID, job.Command, job.State, strconv.Itoa(job.Attempts), job.CreatedAt.Format(time.RFC1123))
	}
	t.Render()
}

func ParseTime(t string) time.Time {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	parsedtime, _ := time.ParseInLocation("2006-01-02 15:04:05", t, loc)
	return parsedtime
}

func PrintStatus(activeWorkers int, jobCountMap map[string]int) {
	clilogger.LogInfo("status")
	t := table.New(os.Stdout)
	t.SetAlignment(table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter)

	t.SetHeaders("💻 workers", "⏳ pending", "🔄 processing", "✅ completed", "❌ failed", "💀 dead")
	t.AddRow(strconv.Itoa(activeWorkers), strconv.Itoa(jobCountMap["pending"]), strconv.Itoa(jobCountMap["processing"]), strconv.Itoa(jobCountMap["completed"]), strconv.Itoa(jobCountMap["failed"]), strconv.Itoa(jobCountMap["dead"]))
	t.Render()
}
