package service

import (
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/kubeproxy/ipvs"
	"minik8s/pkg/kubeproxy/net_interface"
)

type Manager interface {
	CreatSvc(service *core.Service)
	DelSvc(service *core.Service)
	HandlePodModify(pod *core.Pod)
	HandlePodDel(pod *core.Pod)
}

type manager struct {
	services map[types.UID]*core.Service
}

func New() Manager {
	return &manager{
		services: make(map[types.UID]*core.Service),
	}
}

func (m *manager) CreatSvc(service *core.Service) {
	_, found := m.services[service.UID]
	if !found {
		createSvc(service)
		m.services[service.UID] = service
		return
	} else {
		log.Fatalln("service doesn't support update")
	}
}

func (m *manager) DelSvc(service *core.Service) {
	err := ipvs.DelIpvsServices(*service)
	if err != nil {
		log.Println(err)
	}
	err = net_interface.DelIPV4(service.Spec.ClusterIP)
	if err != nil {
		log.Println(err)
	}
}

func (m *manager) HandlePodModify(pod *core.Pod) {
	for label, val := range pod.Labels {
		for _, svc := range m.services {
			v, f := svc.Spec.Selector[label]
			if !f || v != val {
				continue
			}
			err := ipvs.RegPodToService(*svc, *pod)
			if err != nil {
				log.Println(err)
			}
		}
	}

}
func (m *manager) HandlePodDel(pod *core.Pod) {
	for label, val := range pod.Labels {
		for _, svc := range m.services {
			v, f := svc.Spec.Selector[label]
			if !f || v != val {
				continue
			}
			err := ipvs.DelPodToService(*svc, *pod)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func createSvc(service *core.Service) {
	err := ipvs.AddIpvsServices(*service)
	if err != nil {
		log.Println(err)
	}
	err = net_interface.AddIPV4(service.Spec.ClusterIP)
	if err != nil {
		log.Println(err)
	}
}
