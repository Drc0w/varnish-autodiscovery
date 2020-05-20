package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var dContext context.Context
var dClient *client.Client

type DockerManager struct {
	Containers map[string]*DockerData
}

func (dManager *DockerManager) watch(changedChan chan bool) {
	for {
		select {
		case <-time.After(10 * time.Second):
			fmt.Printf("Data check\n")
			if dManager.checkContainerData() {
				fmt.Printf("Data changed\n")
				changedChan <- true
			}
		}
	}
}

func (dManager *DockerManager) checkContainerData() bool {
	newData, err := buildContainers()
	if err != nil {
		// Mask error for now
		return false
	}

	if len(dManager.Containers) != len(newData) {
		dManager.Containers = newData
		return true
	}

	changed := false
	for id := range dManager.Containers {
		changed = !dManager.Containers[id].Equals(newData[id])
		if changed {
			changed = true
			break
		}
	}

	dManager.Containers = newData

	return changed
}

func loadContext() {
	dContext = context.Background()
}

func loadClient() {
	var err error
	dClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	dClient.NegotiateAPIVersion(dContext)
}

func initContext() {
	loadContext()
	loadClient()
}

func buildContainers() (map[string]*DockerData, error) {
	containers, err := dClient.ContainerList(dContext, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	dData := make(map[string]*DockerData)
	for _, container := range containers {
		containerData, err := loadContainerData(container.ID)
		if err != nil {
			continue
		}
		if containerData.ShouldAddBackend() {
			dData[container.ID] = containerData
		}
	}
	return dData, nil
}

func New(dataChangedChannel chan bool) (*DockerManager, error) {
	initContext()
	dData, err := buildContainers()
	if err != nil {
		return nil, err
	}

	dManager := &DockerManager{
		Containers: dData,
	}
	go dManager.watch(dataChangedChannel)

	return dManager, nil
}
