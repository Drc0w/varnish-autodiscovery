package docker

import (
	"errors"
)

func (dData *DockerData) shouldAddServiceBackend() bool {
	if dData == nil {
		return false
	}
	net, err := dData.getServiceIPAddress()
	return err == nil && len(net) > 0
}

func (dData *DockerData) getServiceIPAddress() (string, error) {
	networkName := dData.Labels["network"]
	if len(networkName) != 0 {
		return dData.Networks[networkName].Addr, nil
	}
	return "", errors.New("No network labels found.")
}

func loadServiceData(containerID string) (*DockerData, error) {
	return nil, nil
}

func (dData *DockerData) serviceEquals(cmp *DockerData) bool {
	return true
}

func (dManager *DockerManager) checkServiceData() bool {
	return true
}

func buildServices(map[string]*DockerData) error {
	return nil
}
