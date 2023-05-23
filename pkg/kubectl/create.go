package kubectl

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s/config"
	"net/http"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		switch s {
		case "pod":
			filename := GetFilename()
			url := config.ApiUrl() + "pods/"
			jsonData, err := GetFormJsonData(filename)
			if err != nil {
				fmt.Println("获得格式化数据错误:", err)
				return
			}
			// 创建HTTP请求并设置正文
			req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("创建HTTP请求错误:", err)
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

			// 处理响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}

			fmt.Printf("HTTP响应: %s\n", body)
		case "replicaset":
			filename := GetFilename()
			url := config.ApiUrl() + "replicasets/"
			jsonData, err := GetFormJsonData(filename)
			if err != nil {
				fmt.Println("获得格式化数据错误:", err)
				return
			}
			// 创建HTTP请求并设置正文
			req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("创建HTTP请求错误:", err)
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

			// 处理响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}

			fmt.Printf("HTTP响应: %s\n", body)
		case "hpa":
			filename := GetFilename()
			url := config.ApiUrl() + "hpa/"
			jsonData, err := GetFormJsonData(filename)
			if err != nil {
				fmt.Println("获得格式化数据错误:", err)
				return
			}
			// 创建HTTP请求并设置正文
			req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
			if err != nil {
				fmt.Println("创建HTTP请求错误:", err)
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

			// 处理响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取HTTP响应错误:", err)
				return
			}

			fmt.Printf("HTTP响应: %s\n", body)
		default:
			fmt.Println("请输入正确的参数")
		}

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
