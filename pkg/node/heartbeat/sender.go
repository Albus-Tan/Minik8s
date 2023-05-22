package heartbeat

import (
	"context"
	"fmt"
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/utils"
	"time"
)

type Sender interface {
	Run(ctx context.Context, cancel context.CancelFunc)
}

func NewSender(nodeUID types.UID) Sender {
	hbCli, _ := apiclient.NewRESTClient(types.HeartbeatObjectType)
	return &sender{
		heartbeatClient: hbCli,
		nodeUID:         nodeUID,
		hb:              nil,
	}
}

type sender struct {
	heartbeatClient client.Interface
	nodeUID         types.UID
	hb              *core.Heartbeat
}

func (s *sender) Run(ctx context.Context, cancel context.CancelFunc) {
	log.Printf("[HeartbeatSender] start\n")
	defer log.Printf("[HeartbeatSender] running\n")

	s.initHeartbeat()

	go func() {
		defer cancel()
		defer log.Printf("[HeartbeatSender] finished\n")
		s.periodicallySendHeartbeat(ctx)
	}()
}

func (s *sender) initHeartbeat() {
	s.hb = &core.Heartbeat{
		TypeMeta:   meta.CreateTypeMeta(types.HeartbeatObjectType),
		ObjectMeta: meta.ObjectMeta{},
		Spec: core.HeartbeatSpec{
			NodeUID: s.nodeUID,
		},
		Status: core.HeartbeatStatus{
			HeartbeatID: utils.GenerateHeartbeatID(),
			Timestamp:   time.Now(),
		},
	}

	_, res, err := s.heartbeatClient.Post(s.hb)
	if err != nil {
		panic(fmt.Sprintf("[initHeartbeat] node %v init heartbeat failed\n", s.nodeUID))
		return
	}

	s.hb.UID = res.UID
	s.hb.ResourceVersion = res.ResourceVersion
}

func (s *sender) updateAndSendHeartbeat() error {

	hbItem, err := s.heartbeatClient.Get(s.hb.UID)
	if err != nil {
		log.Printf("[updateAndSendHeartbeat] node %v get heartbeat info failed\n", s.nodeUID)
		return err
	}

	s.hb = hbItem.(*core.Heartbeat)
	s.hb.Status.HeartbeatID = utils.GenerateHeartbeatID()
	s.hb.Status.Timestamp = time.Now()

	_, _, err = s.heartbeatClient.Put(s.hb.UID, s.hb)
	if err != nil {
		log.Printf("[updateAndSendHeartbeat] node %v heartbeat sent failed\n", s.nodeUID)
		return err
	}

	return nil
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
			// send heartbeat by sending heartbeat object to ApiServer
			err := s.updateAndSendHeartbeat()
			if err != nil {
				log.Printf("[periodicallySendHeartbeat] node %v heartbeat sent failed, err: %v\n", s.nodeUID, err)
			} else {
				log.Printf("[periodicallySendHeartbeat] node %v heartbeat sent success\n", s.nodeUID)
			}

			time.Sleep(defaultHeartbeatSendInterval)
		}
	}

}
