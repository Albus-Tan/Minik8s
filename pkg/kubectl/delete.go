package kubectl

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "del <resource> <resource-name>",
	Short: "delete resource",
	Args:  cobra.ExactArgs(2),
	Run:   doDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
