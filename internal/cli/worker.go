package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/utils"
	"github.com/roydevashish/queuectl/internal/worker"
	"github.com/spf13/cobra"
)

var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "control background worker processes that execute queued jobs",
	Long: `The worker subcommand is used to manage workers, providing commands
to start and stop the background processes responsible for consuming
and executing jobs from the queue. 

Use start to launch one or more workers with options for count, and
stop to gracefully or forcefully shut them down.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts one or more worker processes to consume and execute jobs",
	Long: `Launches background worker processes that poll the queue, execute 
jobs concurrently, and handle retries, logging, and graceful shutdown
signals. Each worker runs in its own process/thread depending on the
backend. Workers automatically respect configuration like concurrency
limits, poll intervals, and visibility timeouts.`,

	Run: func(cmd *cobra.Command, args []string) {
		workerCount, _ := cmd.Flags().GetInt("count")
		if workerCount <= 0 {
			clilogger.LogError("no of workers must be greater or equal to 1")
		}

		if running, _ := utils.IsRunning("pid"); running {
			clilogger.LogError("workers already running")
			return
		}

		if err := utils.WriteFile("pid", strconv.Itoa(os.Getpid())); err != nil {
			clilogger.LogError("unable to start workers")
			return
		}

		if err := utils.WriteFile("worker", strconv.Itoa(workerCount)); err != nil {
			clilogger.LogError("unable to write workers count")
			return
		}

		jobChan := make(chan string, 2*workerCount)
		shutdown := make(chan struct{})
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		go worker.StartScheduler(jobChan, shutdown)

		var wg sync.WaitGroup
		for i := 1; i <= workerCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				clilogger.LogInfo(fmt.Sprint("worker started with id: ", id))
				for {
					select {
					case jobID := <-jobChan:
						worker.ExecuteJob(jobID)
					case <-shutdown:
						return
					}
				}
			}(i)
		}

		clilogger.LogInfo(fmt.Sprint("starting total #", workerCount, " workers, press Ctrl+C to stop"))
		<-sig
		clilogger.LogInfo("workers shutting down")
		close(shutdown)
		wg.Wait()
		clilogger.LogSuccess("all workers shutdown")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops all running workers gracefully",
	Long: `Sends SIGTERM to all managed worker processes, allowing in-progress
jobs to complete before shutdown. Workers finish their current job,
commit results, and then exit.`,

	Run: func(cmd *cobra.Command, args []string) {
		running, processId := utils.IsRunning("pid")
		if !running {
			clilogger.LogError("no workers are running right now")
			return
		}

		process, _ := os.FindProcess(processId)
		process.Signal(syscall.SIGINT)
		clilogger.LogSuccess("stoping all workers")
	},
}

func init() {
	WorkerCmd.AddCommand(startCmd, stopCmd)
	startCmd.Flags().IntP("count", "c", 1, "no of workers to start")
}
