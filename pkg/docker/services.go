package docker

import (
	"errors"
	"net"

	"github.com/docker/docker/api/types"
	swarm "github.com/docker/docker/api/types/swarm"
)

// This networks list must be based on IDs and not on names
func loadServiceData(service swarm.Service, networks map[string]*networkData) (*DockerData, error) {
	if service.Endpoint.Spec.Mode != "vip" {
		return nil, errors.New("Not a service with VIP enabled")
	}

	dData := &DockerData{
		ShortID: service.ID[:10],
		ID:      service.ID,
		Name:    service.Spec.Name,
		Labels:  service.Spec.Labels,
		dType:   SERVICE,
	}

	for i := range service.Spec.TaskTemplate.Networks {
		targetID := service.Spec.TaskTemplate.Networks[i].Target
		targetAddr := ""
		for vip := range service.Endpoint.VirtualIPs {
			if service.Endpoint.VirtualIPs[vip].NetworkID == targetID {
				targetAddr = service.Endpoint.VirtualIPs[vip].Addr
				break
			}
		}
		if len(targetAddr) == 0 {
			continue
		}
		ip, _, _ := net.ParseCIDR(targetAddr)
		dData.Networks = make(map[string]*networkData)
		dData.Networks[networks[targetID].Name] = &networkData{
			ID:   targetID,
			Name: networks[targetID].Name,
			Addr: ip.String(),
		}
	}

	return dData, nil
}

func buildServices(dData map[string]*DockerData, networks map[string]*networkData) error {
	services, err := dClient.ServiceList(dContext, types.ServiceListOptions{})
	if err != nil {
		return err
	}

	for _, service := range services {
		serviceData, err := loadServiceData(service, networks)
		if err != nil {
			continue
		}
		if serviceData.shouldAddBackend() {
			dData[service.ID] = serviceData
		}
	}
	return nil
}
