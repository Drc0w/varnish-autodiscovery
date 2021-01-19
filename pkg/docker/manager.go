package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

var dContext context.Context
var dClient *client.Client

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

func (dManager *DockerManager) watch(changedChan chan bool) {
	for {
		select {
		case <-time.After(10 * time.Second):
			fmt.Printf("Data check\n")
			newData, err := build()
			if err != nil {
				panic(err)
			}
			if compareDockerData(dManager.Endpoints, newData) {
				fmt.Printf("Data changed\n")
				dManager.Endpoints = newData
				changedChan <- true
			}
		}
	}
}

func New(dataChangedChannel chan bool) (*DockerManager, error) {
	initContext()
	dData, err := build()
	if err != nil {
		return nil, err
	}

	dManager := &DockerManager{
		Endpoints: dData,
	}
	go dManager.watch(dataChangedChannel)

	return dManager, nil
}
