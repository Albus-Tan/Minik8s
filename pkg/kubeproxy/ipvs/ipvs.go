package ipvs

import (
	"github.com/moby/ipvs"
	"minik8s/pkg/api/core"
	"net"
)

const defaultProtocol = 6
const defaultSchedName = "wlc"
const defaultNetmask = 0xFF_FF_FF_FF
const defaultAddressFamily = 2

func convertToIpvsServices(service core.Service) []ipvs.Service {
	ret := make([]ipvs.Service, 0)
	for _, p := range service.Spec.Ports {
		ret = append(ret, ipvs.Service{
			Address:       net.ParseIP(service.Spec.ClusterIP),
			Protocol:      defaultProtocol,
			Port:          uint16(p.Port),
			SchedName:     defaultSchedName,
			Netmask:       defaultNetmask,
			AddressFamily: defaultAddressFamily,
		})
	}
	return ret
}

func convertToIpvsServiceWithPorts(service core.Service) ([]ipvs.Service, []uint16) {
	retS := make([]ipvs.Service, 0)
	retP := make([]uint16, 0)
	for _, p := range service.Spec.Ports {
		retS = append(retS, ipvs.Service{
			Address:       net.ParseIP(service.Spec.ClusterIP),
			Protocol:      defaultProtocol,
			Port:          uint16(p.Port),
			SchedName:     defaultSchedName,
			Netmask:       defaultNetmask,
			AddressFamily: defaultAddressFamily,
		})
		retP = append(retP, uint16(p.TargetPort))
	}
	return retS, retP
}

func AddIpvsServices(service core.Service) error {
	h, err := ipvs.New("")
	if err != nil {
		return err
	}
	services, err := h.GetServices()
	if err != nil {
		return err
	}
	defer h.Close()
	for _, ipvsSvc := range convertToIpvsServices(service) {
		for _, service := range services {
			if ipvsSvc.Address.Equal(service.Address) && ipvsSvc.Port == service.Port {
				continue
			}
		}

		err := h.NewService(&ipvsSvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func DelIpvsServices(service core.Service) error {
	h, err := ipvs.New("")
	if err != nil {
		return err
	}
	services, err := h.GetServices()
	if err != nil {
		return err
	}
	defer h.Close()
	for _, ipvsSvc := range convertToIpvsServices(service) {
		for _, service := range services {
			if ipvsSvc.Address.Equal(service.Address) && ipvsSvc.Port == service.Port {
				err := h.DelService(&ipvsSvc)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func RegPodToService(service core.Service, pod core.Pod) error {
	h, err := ipvs.New("")
	if err != nil {
		return err
	}
	services, err := h.GetServices()
	if err != nil {
		return err
	}
	defer h.Close()
	ss, ps := convertToIpvsServiceWithPorts(service)
	for idx, ipvsSvc := range ss {
		for _, service := range services {
			if ipvsSvc.Address.Equal(service.Address) && ipvsSvc.Port == service.Port {
				err := h.NewDestination(&ipvsSvc, &ipvs.Destination{
					Address:       net.ParseIP(pod.Status.PodIP),
					Port:          ps[idx],
					AddressFamily: defaultAddressFamily,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func DelPodToService(service core.Service, pod core.Pod) error {
	h, err := ipvs.New("")
	if err != nil {
		return err
	}
	services, err := h.GetServices()
	if err != nil {
		return err
	}
	defer h.Close()
	ss, ps := convertToIpvsServiceWithPorts(service)
	for idx, ipvsSvc := range ss {
		for _, service := range services {
			if ipvsSvc.Address.Equal(service.Address) && ipvsSvc.Port == service.Port {
				err := h.DelDestination(&ipvsSvc, &ipvs.Destination{
					Address:       net.ParseIP(pod.Status.PodIP),
					Port:          ps[idx],
					AddressFamily: defaultAddressFamily,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
