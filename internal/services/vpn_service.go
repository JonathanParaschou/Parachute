package services

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type VPNService struct {
	client *wgctrl.Client
}

func NewVPNService() (*VPNService, error) {
	client, _ := wgctrl.New()
	return &VPNService{client: client}, nil
}

func (v *VPNService) Close() {
	if v.client != nil {
		v.client.Close()
	}
}

func (v *VPNService) CreateInterface(name string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("direct WireGuard interface creation is supported only on Linux")
	}
	if _, err := exec.LookPath("ip"); err != nil {
		return fmt.Errorf("ip command not found; install iproute2 to manage WireGuard interfaces: %w", err)
	}

	// Create WireGuard interface
	cmd := exec.Command("ip", "link", "add", "dev", name, "type", "wireguard")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create WireGuard interface: %w", err)
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up interface: %w", err)
	}

	return nil
}

func (v *VPNService) ConfigureInterface(name string, privateKey wgtypes.Key, listenPort int, peers []wgtypes.PeerConfig) error {
	if v.client == nil {
		return errors.New("WireGuard control client is not available")
	}

	config := wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   &listenPort,
		Peers:        peers,
		ReplacePeers: true,
	}

	if err := v.client.ConfigureDevice(name, config); err != nil {
		return fmt.Errorf("failed to configure WireGuard device: %w", err)
	}

	return nil
}

func (v *VPNService) AssignIP(name, ip string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("WireGuard IP assignment is currently supported only on Linux")
	}
	if _, err := exec.LookPath("ip"); err != nil {
		return fmt.Errorf("ip command not found; install iproute2 to manage WireGuard interfaces: %w", err)
	}

	cmd := exec.Command("ip", "address", "add", ip, "dev", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign IP: %w", err)
	}

	return nil
}

func (v *VPNService) GenerateKeyPair() (wgtypes.Key, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return wgtypes.Key{}, err
	}
	return privateKey, nil
}

func (v *VPNService) GetPublicKey(privateKey wgtypes.Key) wgtypes.Key {
	return privateKey.PublicKey()
}

func (v *VPNService) SetupVPN(name, ipRange string) (*VPNConfig, error) {
	// Generate server keys
	serverPrivateKey, err := v.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	serverPublicKey := v.GetPublicKey(serverPrivateKey)

	serverIP := ipRange + "1/24" // Server gets .1

	config := &VPNConfig{
		InterfaceName:    name,
		ServerPrivateKey: serverPrivateKey,
		ServerPublicKey:  serverPublicKey,
		ServerIP:         serverIP,
		IPRange:          ipRange,
		ListenPort:       51820,
		Peers:            make(map[string]VPNPeer),
	}

	configPath, err := v.WriteServerConfig(config)
	if err != nil {
		return nil, err
	}
	config.ConfigPath = configPath

	if err := v.Activate(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (v *VPNService) WriteServerConfig(config *VPNConfig) (string, error) {
	configDir, err := vpnConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", err
	}

	path := filepath.Join(configDir, config.InterfaceName+".conf")
	if runtime.GOOS == "windows" {
		path = filepath.Join(configDir, config.InterfaceName+".conf")
	}

	body := renderServerConfig(config)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (v *VPNService) Activate(config *VPNConfig) error {
	switch runtime.GOOS {
	case "linux", "darwin":
		return activateWithWGQuick(config.ConfigPath)
	case "windows":
		return activateWithWireGuardExe(config.ConfigPath)
	default:
		return fmt.Errorf("WireGuard activation is not supported on %s", runtime.GOOS)
	}
}

func renderServerConfig(config *VPNConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "PrivateKey = %s\n", config.ServerPrivateKey.String())
	fmt.Fprintf(&b, "Address = %s\n", config.ServerIP)
	fmt.Fprintf(&b, "ListenPort = %d\n", config.ListenPort)

	for name, peer := range config.Peers {
		fmt.Fprintf(&b, "\n[Peer]\n")
		if name != "" {
			fmt.Fprintf(&b, "# %s\n", name)
		}
		fmt.Fprintf(&b, "PublicKey = %s\n", peer.PublicKey.String())
		if len(peer.AllowedIPs) > 0 {
			allowed := make([]string, 0, len(peer.AllowedIPs))
			for _, ip := range peer.AllowedIPs {
				allowed = append(allowed, ip.String())
			}
			fmt.Fprintf(&b, "AllowedIPs = %s\n", strings.Join(allowed, ", "))
		}
		if peer.Endpoint != nil {
			fmt.Fprintf(&b, "Endpoint = %s\n", peer.Endpoint.String())
		}
	}

	return b.String()
}

func activateWithWGQuick(configPath string) error {
	wgQuick, err := exec.LookPath("wg-quick")
	if err != nil {
		return fmt.Errorf("wg-quick not found; install WireGuard tools and rerun setup --vpn: %w", err)
	}

	cmd := exec.Command(wgQuick, "up", configPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wg-quick up failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func activateWithWireGuardExe(configPath string) error {
	wireguard, err := findWireGuardExe()
	if err != nil {
		return err
	}

	cmd := exec.Command(wireguard, "/installtunnelservice", configPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wireguard.exe /installtunnelservice failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func findWireGuardExe() (string, error) {
	if wireguard, err := exec.LookPath("wireguard.exe"); err == nil {
		return wireguard, nil
	}

	for _, candidate := range []string{
		filepath.Join(os.Getenv("ProgramFiles"), "WireGuard", "wireguard.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "WireGuard", "wireguard.exe"),
	} {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", errors.New("wireguard.exe not found; install WireGuard for Windows and rerun setup --vpn as Administrator")
}

func vpnConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "parachute", "wireguard"), nil
}

type VPNConfig struct {
	InterfaceName    string
	ServerPrivateKey wgtypes.Key
	ServerPublicKey  wgtypes.Key
	ServerIP         string
	IPRange          string
	ListenPort       int
	ConfigPath       string
	Peers            map[string]VPNPeer
}

type VPNPeer struct {
	PublicKey  wgtypes.Key
	AllowedIPs []net.IPNet
	Endpoint   *net.UDPAddr
}
