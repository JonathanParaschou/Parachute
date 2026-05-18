package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DashboardPort = 8080

type RemoteAccessService struct{}

type RemoteAccessStatus struct {
	LocalURL       string       `json:"local_url"`
	LANURLs        []string     `json:"lan_urls"`
	Tailscale      RemoteOption `json:"tailscale"`
	Recommended    string       `json:"recommended"`
	ServerPort     int          `json:"server_port"`
	MachineName    string       `json:"machine_name"`
	RemoteWarnings []string     `json:"remote_warnings"`
}

type RemoteOption struct {
	Available  bool   `json:"available"`
	Configured bool   `json:"configured"`
	URL        string `json:"url,omitempty"`
	Address    string `json:"address,omitempty"`
	Message    string `json:"message,omitempty"`
}

func NewRemoteAccessService() *RemoteAccessService {
	return &RemoteAccessService{}
}

func (s *RemoteAccessService) Status() (RemoteAccessStatus, error) {
	name, _ := os.Hostname()
	status := RemoteAccessStatus{
		LocalURL:    dashboardURL("localhost"),
		LANURLs:     lanDashboardURLs(),
		Tailscale:   tailscaleStatus(),
		ServerPort:  DashboardPort,
		MachineName: name,
		Recommended: "local",
	}

	switch {
	case status.Tailscale.Available && status.Tailscale.URL != "":
		status.Recommended = "tailscale"
	case len(status.LANURLs) > 0:
		status.Recommended = "lan"
	}

	if !status.Tailscale.Available {
		status.RemoteWarnings = append(status.RemoteWarnings, status.Tailscale.Message)
	}

	return status, nil
}

func dashboardURL(host string) string {
	return fmt.Sprintf("http://%s:%d", host, DashboardPort)
}

func lanDashboardURLs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var urls []string
	seen := map[string]bool{}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := ipFromAddr(addr)
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			host := ip.String()
			if seen[host] {
				continue
			}
			seen[host] = true
			urls = append(urls, dashboardURL(host))
		}
	}
	return urls
}

func ipFromAddr(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		return nil
	}
}

func tailscaleStatus() RemoteOption {
	tailscale, err := exec.LookPath("tailscale")
	if err != nil {
		if tailscaleExe, exeErr := exec.LookPath("tailscale.exe"); exeErr == nil {
			tailscale = tailscaleExe
		} else {
			return RemoteOption{
				Available: false,
				Message:   "Tailscale CLI not found. Install Tailscale and run tailscale up to enable Tailnet access.",
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, tailscale, "ip", "-4").Output()
	if err != nil {
		msg := "Tailscale is installed, but no Tailscale IPv4 address was reported. Run tailscale up."
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			msg = "Tailscale check timed out."
		}
		return RemoteOption{
			Available: true,
			Message:   msg,
		}
	}

	ip := firstNonEmptyLine(string(out))
	if ip == "" {
		return RemoteOption{
			Available: true,
			Message:   "Tailscale is installed, but no Tailscale IPv4 address was reported. Run tailscale up.",
		}
	}

	return RemoteOption{
		Available:  true,
		Configured: true,
		Address:    ip,
		URL:        dashboardURL(ip),
		Message:    "Tailscale is available for private remote access.",
	}
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
