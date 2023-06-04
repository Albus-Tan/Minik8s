package kubectl

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <resource> -f <filename>",
	Short: "create resources",
	Args:  cobra.MinimumNArgs(1),
	Run:   doCreate,
}

func init() {
	rootCmd.PersistentFlags().StringP("filename", "f", "", "resource name")
	rootCmd.AddCommand(createCmd)
}
