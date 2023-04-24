package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/config"
	"net/http"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := config.ApiUrl() + "watch/" + "pods/"
		s := args[0]
		fmt.Println(s)
		switch s {
		case "pods":
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)

			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "pod":
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url = config.ApiUrl() + "watch/" + "pods/"
			url = url + ":" + podname
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		default:
			fmt.Println("watch error")

		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
