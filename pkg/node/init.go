package node

import (
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"strconv"
)

func CreateWorkerNode() *core.Node {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nc := &NodeCreator{
		nodeClient: nodeCli,
		nodeInfo:   nil,
		ty:         config.Worker,
	}
	nc.initNode()
	nc.registerNode()
	return nc.nodeInfo
}

func CreateMasterNode() *core.Node {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nc := &NodeCreator{
		nodeClient: nodeCli,
		nodeInfo:   nil,
		ty:         config.Master,
	}
	nc.initNode()
	nc.registerNode()
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

func (nc *NodeCreator) initNode() {
	nc.nodeInfo = config.LoadNodeFromTemplate(nc.ty)
	nc.nodeInfo.Name = nc.generateNodeName()
}

var nodeNum int

func init() {
	nodeNum = 0
}

const (
	NameMaster       = "master"
	NameWorkerPrefix = "node"
	NameUndefined    = "undefined"
)

func (nc *NodeCreator) generateNodeName() string {
	var name string
	switch nc.ty {
	case config.Master:
		name = NameMaster
	case config.Worker:
		nodeNum++
		name = NameWorkerPrefix + strconv.Itoa(nodeNum)
	default:
		name = NameUndefined
	}
	log.Printf("[NodeCreator] Generate node name: %v\n", name)
	return name
}

func (nc *NodeCreator) registerNode() {
	_, resp, err := nc.nodeClient.Post(nc.nodeInfo)
	if err != nil {
		panic(err)
		return
	}

	// FIXME: get IP address of physical machine and set address field of node
	nc.nodeInfo.Spec.Address = "localhost"

	nc.nodeInfo.SetUID(resp.UID)
	nc.nodeInfo.SetResourceVersion(resp.ResourceVersion)
	nc.nodeInfo.Status.Phase = core.NodeRunning
	_, _, err = nc.nodeClient.Put(resp.UID, nc.nodeInfo)
	if err != nil {
		panic(err)
		return
	}
}
