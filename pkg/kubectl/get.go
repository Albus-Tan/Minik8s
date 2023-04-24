package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/config"
	"net/http"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		fmt.Println(s)
		switch s {
		case "pods":
			//get localhost:8080/api/pods:name
			url := config.ApiUrl() + "pods/"

			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)

			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "namespaces":
			//get localhost:8080/api/namespaces
			url := config.ApiUrl() + "namespaces/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "pod":
			//get localhost:8080/api/pod:name
			if len(s) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url := config.ApiUrl() + "pods/"
			url = url + ":" + podname
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "podstatus":
			url := config.ApiUrl() + "pods/"
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url = url + ":" + podname + "/status"
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)

		default:
			fmt.Println("get error")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "kube pods' namespace")
	rootCmd.AddCommand(getCmd)
}
