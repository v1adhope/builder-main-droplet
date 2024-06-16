package controllers

import (
	"io"
	"os"
	"path/filepath"
)

type Unbound struct {
	DstConfPath string
}

func NewUnbound(dstConfPath string) *Unbound {
	return &Unbound{dstConfPath}
}

func (u *Unbound) SetUp() error {
	srcConf, err := os.Open(AssetsUnboundConfigPath)
	if err != nil {
		return err
	}
	defer srcConf.Close()

	if err := os.MkdirAll(filepath.Dir(u.DstConfPath), os.ModePerm); err != nil {
		return err
	}

	dstConf, err := os.Create(u.DstConfPath)
	if err != nil {
		return err
	}
	defer dstConf.Close()

	if _, err := io.Copy(dstConf, srcConf); err != nil {
		return err
	}

	return nil
}
