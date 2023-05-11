package cadvisor

import (
	"log"
	"minik8s/config"
	"testing"
)

func TestClient(t *testing.T) {
	cadvisorClient := NewClient(config.CadvisorUrl())
	err := cadvisorClient.Start()
	if err != nil {
		log.Printf("[Test cadvisor] Start cadvisor error: %v\n", err)
		return
	}

	info, err := cadvisorClient.MachineInfo()
	if err != nil {
		log.Printf("[Test cadvisor] MachineInfo error: %v\n", err)
		return
	}
	log.Printf("[Test cadvisor] MachineInfo: %v\n", info)

}
