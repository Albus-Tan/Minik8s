package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s/config"
	"net/http"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all pods.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := config.ApiUrl() + "clear/"
		req, _ := http.NewRequest("GET", url, nil)
		res, _ := http.DefaultClient.Do(req)
		str, _ := io.ReadAll(res.Body)
		fmt.Println(string(str))
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
