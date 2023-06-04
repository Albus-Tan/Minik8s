package kubectl

import (
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <resource> -f <filename>",
	Short: "apply resources",
	Args:  cobra.MinimumNArgs(1),
	Run:   doCreate,
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
