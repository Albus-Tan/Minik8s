package config

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api/core"
	"minik8s/utils"
	"os"
)

/*--------------- Node Basic ---------------*/

type NodeType string

const (
	Master NodeType = "Master"
	Worker NodeType = "Worker"
)

// Name of node in ConfigFile must be unique
// and name of master node must be "master"
func NodeConfig() string {
	return os.Getenv("NODE_CONFIG")
}

func LoadNodeFromTemplate() *core.Node {
	path := NodeConfig()

	file, err := utils.GetFormJsonData(path)
	if err != nil {
		fmt.Printf("Error reading file %s, err %v\n", path, err)
		return nil
	}

	node := &core.Node{}
	err = json.Unmarshal(file, node)
	if err != nil {
		fmt.Println("Error json unmarshal", err)
		return nil
	}

	return node
}
