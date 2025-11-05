package cli

import (
	"fmt"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
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
		clilogger.LogSuccess(fmt.Sprint("start #", workerCount, " workers"))
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops all running workers gracefully",
	Long: `Sends SIGTERM to all managed worker processes, allowing in-progress
jobs to complete before shutdown. Workers finish their current job,
commit results, and then exit.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stop all workers")
	},
}

func init() {
	WorkerCmd.AddCommand(startCmd, stopCmd)
	startCmd.Flags().IntP("count", "c", 1, "no of workers to start")
}
