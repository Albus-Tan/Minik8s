package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/config"
	"net/http"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		switch s {
		case "pod":
			//get localhost:8080/api/pods:name
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url := config.ApiUrl() + "pods/"
			url = url + ":" + podname
			req, _ := http.NewRequest("DELETE", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "namespace":
			//get localhost:8080/api/namespaces
			url := config.ApiUrl() + "namespaces/"
			req, _ := http.NewRequest("DELETE", url, nil)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
