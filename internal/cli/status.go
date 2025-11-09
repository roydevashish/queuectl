package cli

import (
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/roydevashish/queuectl/internal/utils"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "displays a summary of job counts, and active workers",
	Long: `Prints real-time statistics including total jobs in each state
(pending, running, completed, failed, dead) and number of workers.`,

	Run: func(cmd *cobra.Command, args []string) {
		stateCountMap := storage.GetJobCountByState()
		activeWorker := utils.ActiveWorkers()
		utils.PrintStatus(activeWorker, stateCountMap)
	},
}
