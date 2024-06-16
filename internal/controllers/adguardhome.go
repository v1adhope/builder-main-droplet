package controllers

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type AdGuardHome struct {
	DstConfPath string
}

func NewAdGuardHome(dstConfPath string) *AdGuardHome {
	return &AdGuardHome{dstConfPath}
}

func (agh *AdGuardHome) SetUp() error {
	srcConf, err := os.Open(AssetsAdGuardHomeConfigPath)
	if err != nil {
		return err
	}
	defer srcConf.Close()

	if err := os.MkdirAll(filepath.Dir(agh.DstConfPath), os.ModePerm); err != nil {
		return err
	}

	dstConf, err := os.Create(agh.DstConfPath)
	if err != nil {
		return err
	}
	defer dstConf.Close()

	if _, err := io.Copy(dstConf, srcConf); err != nil {
		return err
	}

	return nil
}

func (agh *AdGuardHome) Prepare() error {
	command := "curl -s -S -L https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/master/scripts/install.sh | sh -s -- -v"

	if err := exec.Command("bash", "-c", command).Run(); err != nil {
		return err
	}

	return nil
}

func (agh *AdGuardHome) Enable() error {
	if err := exec.Command("bash", "-c", "/opt/AdGuardHome/AdGuardHome -s start").Run(); err != nil {
		return err
	}

	return nil
}
