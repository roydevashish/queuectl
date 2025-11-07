package utils

import (
	"os"
	"strconv"
	"syscall"
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
