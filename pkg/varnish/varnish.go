package varnish

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"text/template"

	"../docker/"
)

// Default template path
const DefaultTemplatePath = "./default.tpl"

const DefaultVLCPath = "/etc/varnish/default.vcl"

//const DefaultVLCPath = "/dev/stdout"

type VarnishOpts struct {
	TemplatePath string
	VCLPath      string
	ListenPort   string
}

type VarnishWatcher struct {
	TemplateStat os.FileInfo
	Template     *template.Template
}

type VarnishManager struct {
	Opts    VarnishOpts
	command *exec.Cmd
	process *os.Process
	watcher VarnishWatcher
}

func (vManager *VarnishManager) RenderVCL(dData map[string]*vdocker.DockerData) error {
	vManager.InitWatcher()
	f, err := os.OpenFile(vManager.Opts.VCLPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return vManager.watcher.Template.Execute(f, dData)
}

func (vManager *VarnishManager) initCommand() {
	vManager.command = exec.Command(
		"varnishd",
		"-F",
		"-a", ":80",
		"-T", "localhost:6082",
		"-f", "/etc/varnish/default.vcl",
		"-S", "none",
		"-s", "malloc,256m")
}

func (vManager *VarnishManager) Run() {
	vManager.initCommand()
	go func() {
		fmt.Printf("Starting Varnish\n")
		err := vManager.command.Run()
		if err != nil {
			fmt.Printf("Error launching Varnish: %s\n", err.Error())
			panic(err)
		} else {
			fmt.Printf("Varnish exited\n")
			panic(errors.New("Troubles!"))
		}
	}()
}

func (vManager *VarnishManager) CheckTemplateChanged() bool {
	stat, err := os.Stat(vManager.Opts.TemplatePath)
	if err != nil {
		return false
	}
	return stat.Size() != vManager.watcher.TemplateStat.Size() || stat.ModTime() != vManager.watcher.TemplateStat.ModTime()
}

func (vManager *VarnishManager) Stop() {
	vManager.command.Process.Signal(syscall.SIGINT)
}

func (v *VarnishManager) InitWatcher() {
	templatePath := v.Opts.TemplatePath
	v.watcher.Template = template.Must(template.ParseFiles(templatePath))
	v.watcher.TemplateStat, _ = os.Stat(templatePath)
}

func (v *VarnishOpts) SetDefaults() {
	v.TemplatePath = DefaultTemplatePath
	v.VCLPath = DefaultVLCPath
}

func New() *VarnishManager {
	vManager := VarnishManager{}
	vManager.Opts.SetDefaults()
	vManager.initCommand()
	return &vManager
}
