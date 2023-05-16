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
			url = url + podname
			req, _ := http.NewRequest("GET", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)

		case "replicasets":
			url := config.ApiUrl() + "watch/" + "replicasets/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)

		case "replicaset":
			if len(args) < 2 {
				fmt.Println("please input the replicaset name")
				return
			}
			replicasetname := args[1]
			url := config.ApiUrl() + "watch/" + "replicasets/"
			url = url + replicasetname
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "hpas":
			url := config.ApiUrl() + "watch/" + "hpas/"
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			fmt.Println(res)
		case "hpa":
			if len(args) < 2 {
				fmt.Println("please input the hpa name")
				return
			}
			hpaname := args[1]
			url := config.ApiUrl() + "watch/" + "hpas/"
			url = url + hpaname
			req, _ := http.NewRequest("GET", url, nil)
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
