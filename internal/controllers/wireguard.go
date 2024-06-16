package controllers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type WireGuard struct {
	Endpoint          string
	ConfigsDstDirPath string
	KeyPairsCount     int
	KeyPairs          []KeyPair
}

type KeyPair struct {
	Private string
	Public  string
}

type srvTmpl struct {
	PrivateKey string
	Clients    []srvTmplClient
}

type srvTmplClient struct {
	PublicKey string
	Number    int
}

type clientTmpl struct {
	PrivateKey   string
	SrvPublicKey string
	Number       int
	Endpoint     string
}

func NewWireGuard(endpoint, configsDstDirPath string, keyPairsCount int) *WireGuard {
	return &WireGuard{
		Endpoint:          endpoint,
		ConfigsDstDirPath: configsDstDirPath,
		KeyPairsCount:     keyPairsCount,
	}
}

func (wg *WireGuard) SetUp() error {
	if err := wg.generateKeyPairs(); err != nil {
		return err
	}

	if err := wg.parseConfigs(); err != nil {
		return err
	}

	return nil
}

func (wg *WireGuard) generateKeyPairs() error {
	kps, privAlias, pubAlias := make([]KeyPair, 0), "privatekey", "publickey"

	for i := 0; i < wg.KeyPairsCount; i++ {
		cmd := exec.Command(
			"bash",
			"-c",
			fmt.Sprintf("wg genkey | tee %s | wg pubkey | tee %s >> /dev/null", privAlias, pubAlias),
		)

		if err := cmd.Run(); err != nil {
			return err
		}

		privKey, err := wg.extractKeyFromFile(privAlias)
		if err != nil {
			return err
		}

		pubKey, err := wg.extractKeyFromFile(pubAlias)
		if err != nil {
			return err
		}

		kps = append(kps, KeyPair{
			Private: privKey,
			Public:  pubKey,
		})
	}

	if err := wg.cleanKeyFiles(privAlias, pubAlias); err != nil {
		return err
	}

	wg.KeyPairs = kps

	return nil
}

func (wg *WireGuard) cleanKeyFiles(names ...string) error {
	if len(names) == 0 {
		return nil
	}

	for _, name := range names {
		if err := os.Remove(name); err != nil {
			return err
		}
	}

	return nil
}

func (wg *WireGuard) extractKeyFromFile(name string) (string, error) {
	key := ""

	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		key = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return key, nil
}

func (wg *WireGuard) parseConfigs() error {
	if err := os.MkdirAll(filepath.Dir(wg.ConfigsDstDirPath), os.ModePerm); err != nil {
		return err
	}

	srvValues := srvTmpl{PrivateKey: wg.KeyPairs[0].Private}

	clientValues := clientTmpl{
		SrvPublicKey: wg.KeyPairs[0].Public,
		Endpoint:     wg.Endpoint,
	}

	srvClientValues := make([]srvTmplClient, 0)

	for i := 1; i < wg.KeyPairsCount; i++ {
		srvClientValues = append(srvClientValues, srvTmplClient{
			PublicKey: wg.KeyPairs[i].Public,
			Number:    i + 1,
		})

		clientValues.PrivateKey = wg.KeyPairs[i].Private
		clientValues.Number = i + 1

		t := template.Must(
			template.New("client.conf").
				ParseFiles(AssetsWireGuardClientTmplPath))

		file, err := os.Create(fmt.Sprintf("%sclient_%d.conf", wg.ConfigsDstDirPath, i+1))
		if err != nil {
			return err
		}
		defer file.Close()

		// TODO: generate QR Code
		if err := t.Execute(file, clientValues); err != nil {
			return err
		}
	}

	srvValues.Clients = srvClientValues

	t := template.Must(
		template.New("srv.conf").
			ParseFiles(AssetsWireGuardSrvTmplPath))

	file, err := os.Create(fmt.Sprintf("%ssrv.conf", wg.ConfigsDstDirPath))
	if err != nil {
		return err
	}
	defer file.Close()

	if err := t.Execute(file, srvValues); err != nil {
		return err
	}

	return nil
}

func (wg *WireGuard) Enable() error {
	if err := exec.Command("bash", "-c", "systemctl enable wg-quick@srv").Run(); err != nil {
		return err
	}

	return nil
}

func (wg *WireGuard) Prepare() error {
	if err := exec.Command("bash", "-c", "echo \"net.ipv4.ip_forward=1\" >> /etc/sysctl.conf && sysctl -p").Run(); err != nil {
		return err
	}

	if err := exec.Command("bash", "-c", "apt install -y wireguard").Run(); err != nil {
		return err
	}

	return nil
}
