package main

import (
	"fmt"

	"github.com/Drc0w/varnish-autodiscovery/pkg/docker"
	"github.com/Drc0w/varnish-autodiscovery/pkg/varnish"
)

func reload(vManager *varnish.VarnishManager, dManager *docker.DockerManager) error {
	if err := vManager.Reload(dManager); err != nil {
		return err
	}
	fmt.Printf("Configuration reloaded\n")
	return nil
}

func main() {
	dockerChannel := make(chan bool)
	varnishChannel := make(chan bool)

	dManager, err := docker.New(dockerChannel)
	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}

	vManager := varnish.New(varnishChannel)
	err = vManager.RenderVCL(dManager.Endpoints)
	if err != nil {
		panic(err)
	}

	vManager.Run()

	for {
		select {
		case <-dockerChannel:
			if err := reload(vManager, dManager); err != nil {
				panic(err)
			}
		case <-varnishChannel:
			if err := reload(vManager, dManager); err != nil {
				panic(err)
			}
		}
	}
}
