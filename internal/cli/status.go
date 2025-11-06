package cli

import (
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "displays a summary of job counts, and active workers",
	Long: `Prints real-time statistics including total jobs in each state
(pending, running, completed, failed, dead) and number of workers.`,

	Run: func(cmd *cobra.Command, args []string) {
		rows, _ := storage.DB.Query(`SELECT state, COUNT(*) FROM jobs GROUP BY state`)
		defer rows.Close()
		counts := map[string]int{}
		for rows.Next() {
			var state string
			var count int
			rows.Scan(&state, &count)
			counts[state] = count
		}

		// need to fetch active workers
		activeWorkers := strconv.Itoa(4)

		clilogger.LogInfo("status")
		t := table.New(os.Stdout)
		t.SetAlignment(table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter)

		t.SetHeaders("💻 workers", "⏳ pending", "🔄 processing", "✅ completed", "❌ failed", "💀 dead")
		t.AddRow(activeWorkers, strconv.Itoa(counts["pending"]), strconv.Itoa(counts["processing"]), strconv.Itoa(counts["completed"]), strconv.Itoa(counts["failed"]), strconv.Itoa(counts["dead"]))
		t.Render()
	},
}
