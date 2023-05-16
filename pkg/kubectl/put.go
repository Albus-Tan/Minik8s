package kubectl

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s/config"
	"net/http"
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put status or clear.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		switch s {
		case "podstatus":
			if len(args) < 2 {
				fmt.Println("please input the pod name")
				return
			}
			podname := args[1]
			url := config.ApiUrl() + "api/" + "pods/"
			url = url + podname + "/status"
			fileName := GetFilename()
			jsonData, err := GetFormJsonData(fileName)
			if err != nil {
				fmt.Println("get json data error:", err)
				return
			}

			req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("new request error:", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("发送HTTP请求错误:", err)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}
			fmt.Printf("HTTP响应: %s\n", body)
		case "replicasetstatus":
			if len(args) < 2 {
				fmt.Println("please input the replicaset name")
				return
			}
			replicasetname := args[1]
			url := config.ApiUrl() + "api/" + "replicasets/"
			url = url + replicasetname + "/status"
			fileName := GetFilename()
			jsonData, err := GetFormJsonData(fileName)
			if err != nil {
				fmt.Println("get json data error:", err)
				return
			}
			req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("new request error:", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("发送HTTP请求错误:", err)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}
			fmt.Printf("HTTP响应: %s\n", body)
		case "hpastatus":
			if len(args) < 2 {
				fmt.Println("please input the hpa name")
				return
			}
			hpaname := args[1]
			url := config.ApiUrl() + "api/" + "hpas/"
			url = url + hpaname + "/status"
			fileName := GetFilename()
			jsonData, err := GetFormJsonData(fileName)
			if err != nil {
				fmt.Println("get json data error:", err)
				return
			}
			req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("new request error:", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("发送HTTP请求错误:", err)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}
			fmt.Printf("HTTP响应: %s\n", body)
		default:
			fmt.Println("please input the right command")

		}
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
