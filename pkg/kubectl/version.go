package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version subcommand show git version info.",

	Run: func(cmd *cobra.Command, args []string) {
		//output, err := ExecuteCommand("git", "version", args...)
		//if err != nil {
		//	Error(cmd, args, err)
		//}

		fmt.Println("kubectl version", config.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
