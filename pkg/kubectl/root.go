package kubectl

import (
	"errors"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "kubectl is a cli for mini-k8s.",
	Long:  `kubectl is a cli for minik-8s.`,
	Run: func(cmd *cobra.Command, args []string) {
		Error(cmd, args, errors.New("unrecognized command"))
	},
}

func Execute() {
	rootCmd.Execute()
}
