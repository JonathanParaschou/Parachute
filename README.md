# 🪂 ParaChute

**Personal cloud storage on any PC.**

ParaChute is a lightweight, Go-powered cloud storage platform that turns any PC into a secure, always-available cloud. Sync, store, and access your files from anywhere—without complex setup or heavyweight infrastructure.

---

## Features

- 📁 Personal cloud storage on your own hardware
- 🔄 Fast, reliable file sync across devices
- ⚡ Lightweight and efficient (written in Go)
- 🔐 Full data ownership and control
- 🖥️ Runs on Windows, macOS, and Linux
- 🌐 Access your files from anywhere
- 🔒 Built-in WireGuard VPN for secure external access

---

## Installation & Setup

### Prerequisites
- Go 1.25 or later
- Optional VPN support:
  - Linux/macOS: WireGuard tools with `wg-quick` available on `PATH`
  - Windows: WireGuard for Windows with `wireguard.exe` available, run as Administrator

### Building from source

Clone the repository and build the binary:

```bash
git clone <repository-url>
cd parachute
go build ./cmd/parachute
```

This creates a `parachute` executable in the current directory.

### Running directly

You can also run ParaChute without building:

```bash
go run ./cmd/parachute
```

---

## CLI quickstart

ParaChute is moving toward a command line utility that turns a machine into a self-hosted cloud storage node.

Initialize a config:

```bash
parachute setup
```

Set up WireGuard for secure external access:

```bash
parachute setup --vpn
```

ParaChute writes a WireGuard server config under your user config directory and activates it with the native platform tool. Linux and macOS use `wg-quick`; Windows uses `wireguard.exe /installtunnelservice`.

Allocate storage by pointing ParaChute at a directory or drive and choosing the maximum space it may use:

```bash
parachute storage add /path/to/drive --limit 500GB
parachute storage list
```

ParaChute creates a managed `ParachuteStorage` folder containing:

```text
ParachuteStorage/
├── objects/
├── metadata/
├── temp/
└── parachute-root.json
```

Start the local server:

```bash
parachute server start
```

---

## GitHub Pages

This repository includes a static project page in `docs/`.

To publish it with GitHub Pages:

1. Open the repository settings on GitHub.
2. Go to **Pages**.
3. Set **Source** to **Deploy from a branch**.
4. Select the default branch and the `/docs` folder.
5. Save.

GitHub will serve the landing page from `docs/index.html`.
