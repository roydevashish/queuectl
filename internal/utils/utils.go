package utils

import (
	"os"
	"strconv"
	"syscall"
)

func WritePID(file string) error {
	return os.WriteFile(file, []byte(strconv.Itoa(os.Getpid())), 0644)
}

func ReadPID(file string) (int, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(b))
}

func IsRunning(pidFile string) (bool, int) {
	pid, err := ReadPID(pidFile)
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
