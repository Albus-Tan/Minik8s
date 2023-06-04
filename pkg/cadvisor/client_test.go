package cadvisor

import (
	v1 "github.com/google/cadvisor/info/v1"
	"log"
	"minik8s/config"
	"testing"
)

func TestClient(t *testing.T) {
	cadvisorClient := NewClient(config.CadvisorUrl(config.CadvisorHost))
	err := cadvisorClient.Start()
	if err != nil {
		log.Printf("[Test cadvisor] Start cadvisor error: %v\n", err)
		return
	}

	info, err := cadvisorClient.MachineInfo()
	if err != nil {
		log.Printf("[Test cadvisor] MachineInfo error: %#v\n", err)
		return
	}
	log.Printf("[Test cadvisor] MachineInfo: %#v\n", info)

	query := v1.DefaultContainerInfoRequest()

	containerInfos, err := cadvisorClient.AllDockerContainers(&query)
	if err != nil {
		log.Printf("[Test cadvisor] AllDockerContainers error: %#v\n", err)
		return
	}

	for _, containerInfo := range containerInfos {
		log.Printf("[Test cadvisor] AllDockerContainers container id %v Info: %#v\n", containerInfo.Id, containerInfo)
		for _, stat := range containerInfo.Stats {
			log.Printf("[Test cadvisor] Stat: %#v\n", *stat)
		}

	}
}
