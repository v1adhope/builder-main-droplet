package app

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/v1adhope/builder-main-droplet/internal/controllers"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func validateKeyPairCount() (int, error) {
	if len(os.Args) != 2 {
		return -1, errors.New("enter key pair count: make run cc=\"3\"")
	}

	value, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil || value < 2 {
		return -1, err
	}

	return int(value), nil
}

func endpoint() (string, error) {
	resp, err := http.Get("https://ifconfig.me/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func Run() error {
	kpc, err := validateKeyPairCount()
	CheckErr(err)

	unbound := controllers.NewUnbound(controllers.DefaultUnboundDstConfPath)
	CheckErr((unbound.Prepare()))
	CheckErr(unbound.SetUp())
	log.Println("unbound was setup")

	endpoint, err := endpoint()
	CheckErr(err)

	wireguard := controllers.NewWireGuard(endpoint, controllers.DefaultWireGuardAssetsDstDirPath, kpc)
	CheckErr((wireguard.Prepare()))
	CheckErr((wireguard.SetUp()))
	CheckErr((wireguard.Enable()))
	log.Println("wireguard was setup")

	return nil
}
