package docker

import (
	"errors"

	"github.com/docker/docker/api/types"
)

func (dData *DockerData) equals(cmp *DockerData) bool {
	if dData == nil || cmp == nil {
		return dData == cmp
	}

	if dData.ID != cmp.ID || dData.Name != cmp.Name || dData.Labels["network"] != cmp.Labels["network"] || dData.dType != cmp.dType {
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

func (dData *DockerData) shouldAddBackend() bool {
	if dData == nil {
		return false
	}
	net, err := dData.GetIPAddress()
	return err == nil && len(net) > 0
}

func (dData *DockerData) GetIPAddress() (string, error) {
	networkName := dData.Labels["network"]
	if len(networkName) == 0 {
		return "", errors.New("No network labels found.")
	}
	if dData.Networks[networkName] == nil {
		return "", errors.New("No network found for container.")
	}
	return dData.Networks[networkName].Addr, nil
}

func getNetworkList(netData map[string]*networkData) error {
	networks, err := dClient.NetworkList(dContext, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		netData[network.ID] = &networkData{
			Name: network.Name,
			ID:   network.ID,
			Addr: "",
		}
	}
	return nil
}

func build() (map[string]*DockerData, error) {
	networkList := make(map[string]*networkData)
	if err := getNetworkList(networkList); err != nil {
		return nil, err
	}

	dData := make(map[string]*DockerData)
	if err := buildContainers(dData); err != nil {
		return nil, err
	}

	if err := buildServices(dData, networkList); err != nil {
		return nil, err
	}

	return dData, nil
}

func changedDockerData(oldData map[string]*DockerData, newData map[string]*DockerData) bool {
	if len(oldData) != len(newData) {
		return true
	}

	changed := false
	for id := range oldData {
		changed = !oldData[id].equals(newData[id])
		if changed {
			changed = true
			break
		}
	}

	return changed
}
