package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"minik8s/config"
	"net/http"
	"os"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update pods or namespaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := args[0]
		switch s {
		case "pod":
			filename := GetFilename()
			url := config.ApiUrl() + "pods/"

			if len(args) < 2 {
				fmt.Println("please input the pod name")
			}
			podname := args[1]
			url = url + ":" + podname
			// 读取yaml文件
			yamlData, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println("读取文件错误:", err)
				return
			}

			// 反序列化yaml数据
			var data interface{}
			err = yaml.Unmarshal(yamlData, &data)
			if err != nil {
				fmt.Println("反序列化yaml数据错误:", err)
				return
			}

			// 序列化data数据
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("序列化json数据错误:", err)
				return
			}

			// 创建HTTP请求并设置正文
			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
		case "podstatus":
			if len(args) < 2 {
				fmt.Println("please input the pod name")
			}
			podname := args[1]
			url := config.ApiUrl() + "pods/" + ":" + podname + "/status"
			// 读取yaml文件
			filename := GetFilename()
			yamlData, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println("读取文件错误:", err)
				return
			}
			// 反序列化yaml数据
			var data interface{}
			err = yaml.Unmarshal(yamlData, &data)
			if err != nil {
				fmt.Println("反序列化yaml数据错误:", err)
				return
			}
			// 序列化data数据
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("序列化json数据错误:", err)
				return
			}
			// 创建HTTP请求并设置正文
			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
			fmt.Println("暂不支持该资源类型")

		}

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
