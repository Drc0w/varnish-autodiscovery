package main

import (
	"fmt"
	"time"

	"github.com/Drc0w/varnish-autodiscovery/pkg/docker"
	"github.com/Drc0w/varnish-autodiscovery/pkg/varnish"
)

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	dManager, err := docker.New()
	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}

	conf := make(chan map[string]*docker.DockerData)

	vManager := varnish.New()
	err = vManager.RenderVCL(dManager.Containers)
	if err != nil {
		panic(err)
	}

	vManager.Run()

	go func() {
		for {
			if dManager.CheckContainerData() {
				fmt.Printf("Data changed\n")
				conf <- dManager.Containers
			} else if vManager.CheckTemplateChanged() {
				fmt.Printf("Configuration file changed\n")
				conf <- dManager.Containers
			}
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		select {
		case c := <-conf:
			err = vManager.RenderVCL(c)
			if err != nil {
				panic(err)
			}
			err = vManager.Reload()
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}
			fmt.Printf("Configuration reloaded\n")
		}
	}
}
