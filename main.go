package main

import (
	"fmt"
	"time"

	vdocker "./pkg/docker"
	"./pkg/varnish"
)

func CheckContainerData(oldData map[string]*vdocker.DockerData, dData map[string]*vdocker.DockerData) bool {
	if len(oldData) != len(dData) {
		return true
	}
	changed := false
	for id := range oldData {
		changed = !oldData[id].Equals(dData[id])
		if changed {
			return true
		}
	}
	return changed
}

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	vdocker.Init()
	dData, err := vdocker.LoadContainers()
	conf := make(chan map[string]*vdocker.DockerData)

	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}

	vManager := varnish.New()
	err = vManager.RenderVCL(dData)
	if err != nil {
		panic(err)
	}

	vManager.Run()

	go func() {
		oldData := dData
		for {
			dData, err := (vdocker.LoadContainers())
			if err != nil {
				fmt.Printf(err.Error())
				panic(err)
			}
			if CheckContainerData(oldData, dData) {
				fmt.Printf("Data changed\n")
				conf <- dData
			} else if vManager.CheckTemplateChanged() {
				fmt.Printf("Configuration file changed\n")
				conf <- dData
			}
			oldData = dData
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
