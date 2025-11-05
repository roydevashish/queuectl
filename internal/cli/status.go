package cli

import (
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "displays a summary of job counts, and active workers",
	Long: `Prints real-time statistics including total jobs in each state
(pending, running, completed, failed, dead) and number of workers.`,

	Run: func(cmd *cobra.Command, args []string) {
		clilogger.LogSuccess("show system status")
	},
}
