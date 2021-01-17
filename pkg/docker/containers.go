package docker

import (
	"errors"

	"github.com/docker/docker/api/types"
)

func (dData *DockerData) shouldAddContainerBackend() bool {
	if dData == nil {
		return false
	}
	net, err := dData.getContainerIPAddress()
	return err == nil && len(net) > 0
}

func (dData *DockerData) getContainerIPAddress() (string, error) {
	networkName := dData.Labels["network"]
	if len(networkName) != 0 {
		return dData.Networks[networkName].Addr, nil
	}
	return "", errors.New("No network labels found.")
}

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
	if _, err := dData.getContainerIPAddress(); err != nil {
		return nil, err
	}
	return dData, nil
}

func (dData *DockerData) containerEquals(cmp *DockerData) bool {
	if dData == nil || cmp == nil {
		return dData == cmp
	}

	if dData.ID != cmp.ID || dData.Name != cmp.Name || dData.Labels["network"] != cmp.Labels["network"] {
		return false
	}

	netChanged := false
	for net := range dData.Networks {
		if netChanged {
			break
		}
		netChanged = cmp.Networks[net] == nil ||
			dData.Networks[net].ID != cmp.Networks[net].ID ||
			dData.Networks[net].Name != cmp.Networks[net].Name ||
			dData.Networks[net].Addr != cmp.Networks[net].Addr
	}
	for net := range cmp.Networks {
		if netChanged {
			break
		}
		netChanged = dData.Networks[net] == nil ||
			cmp.Networks[net].ID != dData.Networks[net].ID ||
			cmp.Networks[net].Name != dData.Networks[net].Name ||
			cmp.Networks[net].Addr != dData.Networks[net].Addr
	}

	return !netChanged
}

func (dManager *DockerManager) checkContainerData() bool {
	newData := make(map[string]*DockerData)
	err := buildContainers(newData)
	if err != nil {
		// Mask error for now
		return false
	}

	if len(dManager.Endpoints) != len(newData) {
		dManager.Endpoints = newData
		return true
	}

	changed := false
	for id := range dManager.Endpoints {
		changed = !dManager.Endpoints[id].containerEquals(newData[id])
		if changed {
			changed = true
			break
		}
	}

	dManager.Endpoints = newData

	return changed
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
		if containerData.shouldAddContainerBackend() {
			dData[container.ID] = containerData
		}
	}
	return nil
}
