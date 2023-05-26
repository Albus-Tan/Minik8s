package node

import (
	"fmt"
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/utils"
	"reflect"
)

func CreateWorkerNode(configFileName string) *core.Node {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nc := &NodeCreator{
		nodeClient: nodeCli,
		nodeInfo:   nil,
		ty:         config.Worker,
	}
	if nc.initNode(configFileName) {
		nc.registerNode()
	}
	return nc.nodeInfo
}

func CreateMasterNode() *core.Node {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nc := &NodeCreator{
		nodeClient: nodeCli,
		nodeInfo:   nil,
		ty:         config.Master,
	}
	if nc.initNode(config.MasterNodeConfigFileName) {
		nc.registerNode()
	}

	return nc.nodeInfo
}

func DeleteNode(n *core.Node) {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	_, _, err := nodeCli.Delete(n.UID)
	if err != nil {
		panic(err)
		return
	}
}

type NodeCreator struct {
	ty         config.NodeType
	nodeClient client.Interface
	nodeInfo   *core.Node
}

func (nc *NodeCreator) initNode(configFileName string) bool {

	nc.nodeInfo = config.LoadNodeFromTemplate(configFileName)

	// FIXME: get IP address of physical machine and set address field of node
	nc.nodeInfo.Spec.Address = "localhost"

	if nc.ty == config.Master {
		nc.nodeInfo.Name = NameMaster
	}
	if nc.nodeInfo.Name == NameEmpty {
		nc.nodeInfo.Name = nc.generateNodeName()
	} else {
		// check if node name exist
		nodeList, err := nc.nodeClient.GetAll()
		if err != nil {
			panic(err)
			return false
		}
		nodeItems := nodeList.GetIApiObjectArr()
		for _, nodeItem := range nodeItems {
			n := nodeItem.(*core.Node)
			if n.Name == nc.nodeInfo.Name {
				if reflect.DeepEqual(n.Spec, nc.nodeInfo.Spec) && reflect.DeepEqual(n.Labels, nc.nodeInfo.Labels) && reflect.DeepEqual(n.TypeMeta, nc.nodeInfo.TypeMeta) {
					// node same with before
					log.Printf("[initNode] config same with node(name %v uid %v) before, reuse it\n", n.Name, n.UID)
					nc.nodeInfo = n
					return false
				}
				if nc.ty == config.Master {
					// master node exist
					panic("master exist, can not create another")
				} else {
					// node name exist
					panic(fmt.Sprintf("node name %v exist, can not create another", nc.nodeInfo.Name))
				}
			}
		}
	}
	return true
}

const (
	NameMaster       = "master"
	NameWorkerPrefix = "node"
	NameUndefined    = "undefined"
	NameEmpty        = ""
)

func (nc *NodeCreator) generateNodeName() string {
	var name string
	switch nc.ty {
	case config.Master:
		name = NameMaster
	case config.Worker:
		name = utils.AppendRandomNameSuffix(NameWorkerPrefix)
	default:
		name = NameUndefined
	}
	log.Printf("[NodeCreator] Generate node name: %v\n", name)
	return name
}

func (nc *NodeCreator) registerNode() {
	nc.nodeInfo.Status.Phase = core.NodePending
	_, resp, err := nc.nodeClient.Post(nc.nodeInfo)
	if err != nil {
		panic(err)
		return
	}
	nc.nodeInfo.SetUID(resp.UID)
	nc.nodeInfo.SetResourceVersion(resp.ResourceVersion)
	nc.nodeInfo.Status.Phase = core.NodeRunning
	_, _, err = nc.nodeClient.Put(resp.UID, nc.nodeInfo)
	if err != nil {
		panic(err)
		return
	}
}
