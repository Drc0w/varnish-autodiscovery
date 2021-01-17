package docker

func (dData *DockerData) GetIPAddress() (string, error) {
	if dData.dType == CONTAINER {
		return dData.getContainerIPAddress()
	} else {
		return dData.getServiceIPAddress()
	}
}

func build() (map[string]*DockerData, error) {
	dData := make(map[string]*DockerData)

	if err := buildContainers(dData); err != nil {
		return nil, err
	}

	err := buildServices(dData)
	if err != nil {
		return nil, err
	}

	return dData, nil
}
