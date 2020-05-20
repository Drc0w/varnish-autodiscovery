package vdocker

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var dContext context.Context
var dClient *client.Client

type DockerData struct {
	ShortID  string
	ID       string
	Name     string
	Networks map[string]*networkData
	Labels   map[string]string
}

type networkData struct {
	ID   string
	Name string
	Addr string
}

func (dData *DockerData) ShouldAddBackend() bool {
	if dData == nil {
		return false
	}
	net, err := dData.GetIPAddressFromNetwork()
	return err == nil && len(net) > 0
}

func (dData *DockerData) GetIPAddressFromNetwork() (string, error) {
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
	if _, err := dData.GetIPAddressFromNetwork(); err != nil {
		return nil, err
	}
	return dData, nil
}

func (dData *DockerData) Equals(rightData *DockerData) bool {
	if (dData == nil && rightData != nil) || (dData != nil && rightData == nil) {
		return true
	}
	netChanged := false
	labelChanged := dData.Labels["network"] != rightData.Labels["network"]

	for net := range dData.Networks {
		if netChanged {
			break
		}
		netChanged = rightData.Networks[net] == nil ||
			dData.Networks[net].ID != rightData.Networks[net].ID ||
			dData.Networks[net].Name != rightData.Networks[net].Name ||
			dData.Networks[net].Addr != rightData.Networks[net].Addr
	}
	for net := range rightData.Networks {
		if netChanged {
			break
		}
		netChanged = dData.Networks[net] == nil ||
			rightData.Networks[net].ID != dData.Networks[net].ID ||
			rightData.Networks[net].Name != dData.Networks[net].Name ||
			rightData.Networks[net].Addr != dData.Networks[net].Addr
	}

	return dData.ID == rightData.ID && dData.Name == rightData.Name && !netChanged && !labelChanged
}

func LoadContainers() (map[string]*DockerData, error) {
	containers, err := dClient.ContainerList(dContext, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	dData := make(map[string]*DockerData)
	for _, container := range containers {
		containerData, err := loadContainerData(container.ID)
		if containerData.ShouldAddBackend() {
			dData[container.ID] = containerData
		}
		if err != nil {
			continue
		}
	}
	return dData, nil
}

func LoadContext() {
	dContext = context.Background()
}

func LoadClient() {
	var err error
	dClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	dClient.NegotiateAPIVersion(dContext)
}

func Init() {
	LoadContext()
	LoadClient()
}
