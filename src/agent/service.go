package main

import (
	"bytes"
	"fmt"
	"github.com/beyond-net/golib/logs"
	"os"
	"os/exec"
	"text/template"
	"time"
)

const (
	gtunConfigFile     = "/opt/apps/gtun/gtun.yaml"
	gtunConfigTemplate = "../etc/gtun.template"
	gtunService        = "gtun.service"
)

type GtunConfigItem struct {
	Region          string
	Scheme          string
	ServerIP        string
	ServerTracePort int
	ServerPort      int
	ListenPort      int
	Rate            int
}

func ReloadGtun(config []*GtunConfigItem) error {
	// render config
	tpl, err := template.ParseFiles(gtunConfigTemplate)
	if err != nil {
		return err
	}

	br := &bytes.Buffer{}
	err = tpl.Execute(br, config)
	if err != nil {
		return err
	}

	return ReloadService(gtunConfigFile, br.String(), gtunService)
}

func ReloadService(configFile, configContent, service string) error {
	// backup config file
	_, err := os.Stat(configFile)
	reloadSuccess := false
	if err == nil {
		backupFile := fmt.Sprintf("%s.%d", gtunConfigFile, time.Now().UnixMicro())
		_, err = exec.Command("cp", []string{gtunConfigFile, backupFile}...).CombinedOutput()
		if err != nil {
			return err
		}

		// recover
		defer func() {
			if !reloadSuccess {
				_, err := exec.Command("cp", []string{backupFile, gtunConfigFile}...).CombinedOutput()
				if err != nil {
					logs.Error("recover config file fail: %v", err)
				}
			}
		}()
	}

	// write new config
	fp, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.Write([]byte(configContent))
	if err != nil {
		return err
	}

	// restart service
	err = RestartService(service)
	if err != nil {
		return err
	}
	reloadSuccess = true
	return nil
}

func RestartService(service string) error {
	_, err := exec.Command("systemctl", []string{"restart", service}...).CombinedOutput()
	return err
}
