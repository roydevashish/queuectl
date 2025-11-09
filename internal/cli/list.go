package cli

import (
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/roydevashish/queuectl/internal/utils"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "lists jobs filtered by state",
	Long: `Lists jobs matching the criteria and displays them in a table with 
columns: ID, Command, State, Attempts, Created At.`,

	Run: func(cmd *cobra.Command, args []string) {
		state, _ := cmd.Flags().GetString("state")

		jobs, err := storage.GetJobListFilterByState(state)
		if err != nil {
			clilogger.LogError(err.Error())
		}
		utils.PrintJobs(jobs)
	},
}

func init() {
	ListCmd.Flags().StringP("state", "s", "", "filter by state: pending, running, completed, failed, dead")
}
