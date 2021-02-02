package docker

import (
	"errors"

	"github.com/docker/docker/api/types"
)

func loadContainerData(containerID string) (*DockerData, error) {
	rawData, err := dClient.ContainerInspect(dContext, containerID)
	if err != nil {
		return nil, err
	}

	dData := &DockerData{
		ShortID: containerID[:10],
		ID:      containerID,
		Name:    rawData.Name,
		Labels:  rawData.Config.Labels,
		dType:   CONTAINER,
	}

	if rawData.NetworkSettings != nil && rawData.NetworkSettings.Networks != nil {
		dData.Networks = make(map[string]*networkData)
		for name, network := range rawData.NetworkSettings.Networks {
			dData.Networks[name] = &networkData{
				ID:   network.NetworkID,
				Name: name,
				Addr: network.IPAddress,
			}
		}
	} else {
		return nil, errors.New("No network settings found.")
	}
	if _, err := dData.GetIPAddress(); err != nil {
		return nil, err
	}
	return dData, nil
}

func buildContainers(dData map[string]*DockerData) error {
	containers, err := dClient.ContainerList(dContext, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		containerData, err := loadContainerData(container.ID)
		if err != nil {
			continue
		}
		if containerData.shouldAddBackend() {
			dData[container.ID] = containerData
		}
	}
	return nil
}
