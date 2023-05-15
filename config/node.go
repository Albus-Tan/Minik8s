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

const (
	RelativePath             = "config"
	MasterNodeConfigFileName = "master.json"
	WorkerNodeConfigFileName = "worker.json"
)

func LoadNodeFromTemplate(t NodeType) *core.Node {
	path, _ := os.Getwd()
	path = filepath.Join(path, RelativePath)
	switch t {
	case Master:
		path = filepath.Join(path, MasterNodeConfigFileName)
	case Worker:
		path = filepath.Join(path, WorkerNodeConfigFileName)
	default:
		return nil
	}

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
