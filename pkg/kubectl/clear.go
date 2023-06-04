package kubectl

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s/config"
	"minik8s/pkg/api"
	"net/http"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all resources",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := config.ApiServerUrl() + "/clear/"
		req, _ := http.NewRequest("GET", url, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("clear request sent failed, err:", err)
			return
		}
		str, _ := io.ReadAll(res.Body)
		var ifErr api.Response
		err = json.Unmarshal(str, &ifErr)
		if err == nil && ifErr.Status == "ERR" {
			fmt.Println("clear failed, err: ", ifErr.ErrorMsg)
			return
		}
		fmt.Println("clear success")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
