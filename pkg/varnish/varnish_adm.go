package varnish

import (
	"errors"
	"net"
	"strings"
	"time"
)

const BUFFER_SIZE = 1024

func sendCmd(conn *net.Conn, cmd string) error {
	_, err := (*conn).Write([]byte(cmd + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func checkCmd(conn *net.Conn) error {
	respBuffer := make([]byte, BUFFER_SIZE)
	respLen, err := (*conn).Read(respBuffer)
	if err != nil {
		return err
	}
	if respLen < 3 || string(respBuffer)[:3] != "200" {
		return errors.New(strings.SplitAfterN(string(respBuffer), "\n", 2)[1])
	}
	return nil
}

func handleCmd(conn *net.Conn, cmd string) error {
	if err := sendCmd(conn, cmd); err != nil {
		return err
	}
	if err := checkCmd(conn); err != nil {
		return err
	}
	return nil
}

func generateVCLName() string {
	uniqueId := time.Now().Format("20060102150405")
	vclName := "vcl_" + uniqueId
	return vclName
}

func openConnection() (*net.Conn, error) {
	var d net.Dialer
	conn, err := d.Dial("tcp", "127.0.0.1:6082")
	if err != nil {
		return nil, err
	}

	if err := checkCmd(&conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &conn, nil
}

func compileVCL(conn *net.Conn, vclName string, vclPath string) error {
	cmd := "vcl.load " + vclName + " " + vclPath
	return handleCmd(conn, cmd)
}

func useVCL(conn *net.Conn, vclName string) error {
	cmd := "vcl.use " + vclName
	return handleCmd(conn, cmd)
}

func (vManager *VarnishManager) Reload() error {
	vclName := generateVCLName()

	conn, err := openConnection()
	if err != nil {
		return err
	}
	defer (*conn).Close()

	// Load new VCL
	if err := compileVCL(conn, vclName, vManager.Opts.VCLPath); err != nil {
		return err
	}

	// Make new VCL effective
	if err := useVCL(conn, vclName); err != nil {
		return err
	}

	return nil
}
