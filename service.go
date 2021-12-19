package main

import (
	"bytes"
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type ServiceMan struct {
	Name        string `flag:"name" short:"name" usage:"服务ID"`
	Description string `flag:"description" short:"description" usage:"服务描述"`
	Command     string `flag:"command" short:"command" usage:"执行命令"`
}

func (s *ServiceMan) control(args ...string) error {
	output, err := exec.Command("systemctl", args...).CombinedOutput()
	if len(output) > 0 {
		Standard("服务").Infof("%s", output)
	}
	return err
}

func (s *ServiceMan) getServiceFilepath() string {
	return filepath.Join("/etc/systemd/system", s.Name+".service")
}

// 安装服务
func (s *ServiceMan) Install() error {
	var buf bytes.Buffer
	if err := template.Must(template.New("").Parse(serviceTemplate)).Execute(&buf, s); err != nil {
		return err
	}

	if err := os.WriteFile(s.getServiceFilepath(), buf.Bytes(), 0666); err != nil {
		return err
	}

	if err := s.control("daemon-reload"); err != nil {
		return err
	}
	return s.control("enable", s.Name)
}

// 卸载服务
func (s *ServiceMan) Uninstall() error {
	if err := s.Stop(); err != nil {
		return err
	}
	if err := os.Remove(s.getServiceFilepath()); err != nil {
		return err
	}
	return s.control("daemon-reload")
}

// 启动服务
func (s *ServiceMan) Start() error {
	return s.control("start", s.Name)
}

// 服务状态
func (s *ServiceMan) Status() error {
	return s.control("status", s.Name)
}

// 停止服务
func (s *ServiceMan) Stop() error {
	return s.control("stop", s.Name)
}

var (
	//go:embed template.service
	serviceTemplate string
	_               embed.FS
)
