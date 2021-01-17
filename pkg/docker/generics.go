package docker

type dType int

const (
	CONTAINER = 1
	SERVICE   = 2
)

type DockerManager struct {
	Endpoints map[string]*DockerData
}

type DockerData struct {
	ShortID  string
	ID       string
	Name     string
	Networks map[string]*networkData
	Labels   map[string]string
	dType    dType
}

type networkData struct {
	ID   string
	Name string
	Addr string
}
