package config

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api/core"
	"os"
	"path/filepath"
)

/*--------------- Node Basic ---------------*/

type NodeType string

const (
	Master NodeType = "Master"
	Worker NodeType = "Worker"
)

// Name of node in ConfigFile must be unique
// and name of master node must be "master"
const (
	RelativePath              = "config"
	MasterNodeConfigFileName  = "master.json"
	Worker1NodeConfigFileName = "worker1.json"
	Worker2NodeConfigFileName = "worker2.json"
)

func LoadNodeFromTemplate(configFileName string) *core.Node {
	path, _ := os.Getwd()
	path = filepath.Join(path, RelativePath)
	path = filepath.Join(path, configFileName)

	file, err := os.ReadFile(path)
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
