package heartbeat

import (
	"context"
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"time"
)

type Sender interface {
	Run(ctx context.Context, cancel context.CancelFunc)
}

func NewSender(nodeUID types.UID) Sender {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	return &sender{
		nodeClient: nodeCli,
		nodeUID:    nodeUID,
	}
}

type sender struct {
	nodeClient client.Interface
	nodeUID    types.UID
}

func (s *sender) Run(ctx context.Context, cancel context.CancelFunc) {
	log.Printf("[HeartbeatSender] start\n")
	defer log.Printf("[HeartbeatSender] running\n")

	go func() {
		defer cancel()
		defer log.Printf("[HeartbeatSender] finished\n")
		s.periodicallySendHeartbeat(ctx)
	}()
}

const defaultHeartbeatSendInterval = config.HeartbeatInterval

func (s *sender) periodicallySendHeartbeat(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[periodicallySendHeartbeat] ctx.Done() received, heartbeat sender exit\n")
			return
		default:
			// send heartbeat by updating node info

			nodeItem, err := s.nodeClient.Get(s.nodeUID)
			if err != nil {
				log.Printf("[periodicallySendHeartbeat] node %v get info failed\n", s.nodeUID)
				continue
			}

			node := nodeItem.(*core.Node)

			_, _, err = s.nodeClient.Put(s.nodeUID, node)
			if err != nil {
				log.Printf("[periodicallySendHeartbeat] node %v heartbeat sent failed\n", s.nodeUID)
				continue
			}

			log.Printf("[periodicallySendHeartbeat] node %v heartbeat sent success\n", s.nodeUID)

			time.Sleep(defaultHeartbeatSendInterval)
		}
	}

}
