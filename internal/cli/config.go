package cli

import (
	"fmt"

	clilogger "github.com/roydevashish/queuectl/internal/cli_logger"
	"github.com/roydevashish/queuectl/internal/storage"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "modify configuration settings such as max retries, and backoff policies",
	Long: `The config subcommand allows management of queue behavior.

Settings include global defaults like max-retries, base_backoff.
Changes are persisted to the config backend and take effect on
next worker startup.`,
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "updates runtime configuration values",
	Long: `Modifies persistent config (stored in file or backend). 
Common keys: max-retries, base_backoff. 
Changes take effect on next worker start.`,

	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]

		_, err := storage.DB.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, value)
		if err != nil {
			clilogger.LogError("unable to set configuration")
			return
		}
		clilogger.LogSuccess(fmt.Sprint("set config ", key, ":", value))
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
