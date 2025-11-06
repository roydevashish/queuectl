package cli

import (
	"os"
	"strconv"

	"github.com/aquasecurity/table"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "lists jobs filtered by state",
	Long: `Lists jobs matching the criteria and displays them in a table with 
columns: ID, Command, State, Attempts, Created At.`,

	Run: func(cmd *cobra.Command, args []string) {
		state, _ := cmd.Flags().GetString("state")
		query := `SELECT id, command, state, attempts, created_at FROM jobs`
		if state != "" {
			query += ` WHERE state = '` + state + `'`
		}
		query += ` ORDER BY created_at DESC LIMIT 20`

		rows, _ := storage.DB.Query(query)
		defer rows.Close()

		t := table.New(os.Stdout)
		t.SetHeaders("id", "command", "state", "attempts", "created at")

		for rows.Next() {
			var id, command, state string
			var attempts int
			var created string
			rows.Scan(&id, &command, &state, &attempts, &created)
			t.AddRow(id, command, state, strconv.Itoa(attempts), created)
		}

		t.Render()
	},
}

func init() {
	ListCmd.Flags().StringP("state", "s", "", "filter by state: pending, running, completed, failed, dead")
}
