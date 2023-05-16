package kubectl

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// GetFormJsonData Get formed json data from a yaml file or a json file
func GetFormJsonData(filename string) (jsonData []byte, err error) {
	// if filename has suffix .yaml or .yml
	if filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml" {
		// 读取yaml文件
		yamlData, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("读取文件错误:", err)
			return nil, err
		}

		// 反序列化yaml数据
		var data interface{}
		err = yaml.Unmarshal(yamlData, &data)
		if err != nil {
			fmt.Println("反序列化yaml数据错误:", err)
			return nil, err
		}

		// 序列化data数据
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("序列化json数据错误:", err)
			return nil, err
		}
		return jsonData, nil
	}
	// if filename has suffix .json
	if filename[len(filename)-5:] == ".json" {
		// 读取json文件
		jsonData, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("读取文件错误:", err)
			return nil, err
		}
		return jsonData, nil

	}

	err = fmt.Errorf("the file %s is not a yaml or json file", filename)
	return nil, err

}

func ExecuteCommand(name string, subname string, args ...string) (string, error) {
	args = append([]string{subname}, args...)

	cmd := exec.Command(name, args...)
	bytes, err := cmd.CombinedOutput()

	return string(bytes), err
}

func GetNamespace() string {
	namespace := "default"
	name, err := rootCmd.PersistentFlags().GetString("namespace")
	if err != nil {
		fmt.Println("the err is", err)
	}
	if name != "" {
		namespace = name
	}
	return namespace
}

func GetFilename() string {
	filename := ""
	name, err := rootCmd.PersistentFlags().GetString("filename")
	if err != nil {
		fmt.Println("the err is", err)
	}
	if name != "" {
		filename = name
	}
	return filename
}

func Error(cmd *cobra.Command, args []string, err error) {
	fmt.Fprintf(os.Stderr, "execute %s args:%v error:%v\n", cmd.Name(), args, err)
	os.Exit(1)
}
