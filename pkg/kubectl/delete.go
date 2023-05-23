package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
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
			//delete localhost:8080/api/pods:name
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url := config.ApiUrl() + "pods/"
			url = url + podname
			req, _ := http.NewRequest("DELETE", url, nil)
			namespace := GetNamespace()
			req.Header.Add("namespace", namespace)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			fmt.Println(string(str))
		case "namespace":
			//delete localhost:8080/api/namespaces
			url := config.ApiUrl() + "namespaces/"
			req, _ := http.NewRequest("DELETE", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			fmt.Println(string(str))
		case "node":
			//delete localhost:8080/api/nodes/:name
			if len(args) < 2 {
				fmt.Println("please input the node name")
				return
			}
			nodename := args[1]
			url := config.ApiUrl() + "nodes/"
			url = url + nodename
			req, _ := http.NewRequest("DELETE", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			fmt.Println(string(str))

		case "replicaset":
			//delete localhost:8080/api/replicasets/:name
			if len(args) < 2 {
				fmt.Println("please input the replicaset name")
				return
			}
			replicasetname := args[1]
			url := config.ApiUrl() + "replicasets/"
			url = url + replicasetname
			req, _ := http.NewRequest("DELETE", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			fmt.Println(string(str))
		case "hpa":
			//delete localhost:8080/api/hpa/:name
			if len(args) < 2 {
				fmt.Println("please input the hpa name")
				return
			}
			hpaname := args[1]
			url := config.ApiUrl() + "hpa/"
			url = url + hpaname
			req, _ := http.NewRequest("DELETE", url, nil)
			res, _ := http.DefaultClient.Do(req)
			str, _ := io.ReadAll(res.Body)
			fmt.Println(string(str))
		default:
			fmt.Println("please input the right command")
		}

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
