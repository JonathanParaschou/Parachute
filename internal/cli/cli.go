package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"parachute/internal/app"
	"parachute/internal/config"
	"parachute/internal/services"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	case "setup":
		return runSetup(args[1:], stdout, stderr)
	case "storage":
		return runStorage(args[1:], stdout, stderr)
	case "server":
		return runServer(args[1:], stdout, stderr)
	case "remote":
		return runRemote(args[1:], stdout, stderr)
	case "start":
		return runServer([]string{"start"}, stdout, stderr)
	case "status":
		return runStatus(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func runSetup(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	path := fs.String("path", "", "storage location to initialize")
	limit := fs.String("limit", "100GB", "maximum space ParaChute may use")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "setup failed: %v\n", err)
		return 1
	}

	if *path == "" {
		cfgPath, err := config.Path()
		if err != nil {
			fmt.Fprintf(stderr, "config path: %v\n", err)
			return 1
		}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(stderr, "setup failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Created config at %s\n", cfgPath)
		fmt.Fprintln(stdout, "Add storage with: parachute storage add <path> --limit 500GB")
		return 0
	}

	return runStorageAdd([]string{*path, "--limit", *limit}, stdout, stderr)
}

func runStorage(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printStorageHelp(stderr)
		return 2
	}

	switch args[0] {
	case "add":
		return runStorageAdd(args[1:], stdout, stderr)
	case "list", "ls":
		return runStorageList(stdout, stderr)
	case "remove", "rm":
		return runStorageRemove(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown storage command: %s\n\n", args[0])
		printStorageHelp(stderr)
		return 2
	}
}

func runStorageAdd(args []string, stdout, stderr io.Writer) int {
	path, limit, err := parseStorageAddArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		fmt.Fprintln(stderr, "usage: parachute storage add <path> --limit 500GB")
		return 2
	}

	limitBytes, err := parseSize(limit)
	if err != nil {
		fmt.Fprintf(stderr, "invalid limit: %v\n", err)
		return 2
	}

	root, err := config.AddStorageRoot(path, limitBytes)
	if err != nil {
		fmt.Fprintf(stderr, "storage add failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Added storage root %s\n", root.Path)
	fmt.Fprintf(stdout, "Limit: %s\n", formatSize(root.LimitBytes))
	return 0
}

func parseStorageAddArgs(args []string) (path string, limit string, err error) {
	limit = "100GB"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--limit":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--limit requires a value")
			}
			limit = args[i+1]
			i++
		default:
			if path != "" {
				return "", "", fmt.Errorf("only one storage path may be provided")
			}
			path = arg
		}
	}
	if path == "" {
		return "", "", fmt.Errorf("storage path is required")
	}
	return path, limit, nil
}

func runStorageList(stdout, stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "storage list failed: %v\n", err)
		return 1
	}
	if len(cfg.StorageRoots) == 0 {
		fmt.Fprintln(stdout, "No storage roots configured.")
		return 0
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tLIMIT\tENABLED\tPATH")
	for _, root := range cfg.StorageRoots {
		fmt.Fprintf(tw, "%s\t%s\t%t\t%s\n", root.ID, formatSize(root.LimitBytes), root.Enabled, root.Path)
	}
	tw.Flush()
	return 0
}

func runStorageRemove(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: parachute storage remove <path>")
		return 2
	}

	root, err := config.RemoveStorageRoot(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "storage remove failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Removed storage root %s\n", root.Path)
	fmt.Fprintln(stdout, "Data on disk was left untouched.")
	return 0
}

func runServer(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "start" {
		fmt.Fprintln(stderr, "usage: parachute server start")
		return 2
	}
	fmt.Fprintln(stdout, "Starting ParaChute server on http://localhost:8080")
	if err := app.Run(); err != nil {
		fmt.Fprintf(stderr, "server failed: %v\n", err)
		return 1
	}
	return 0
}

func runRemote(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "status" {
		fmt.Fprintln(stderr, "usage: parachute remote status")
		return 2
	}

	remoteAccessService := services.NewRemoteAccessService()
	status, err := remoteAccessService.Status()
	if err != nil {
		fmt.Fprintf(stderr, "remote status failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Dashboard port: %d\n", status.ServerPort)
	fmt.Fprintf(stdout, "Local: %s\n", status.LocalURL)
	if len(status.LANURLs) == 0 {
		fmt.Fprintln(stdout, "LAN: no LAN IPv4 address detected")
	} else {
		fmt.Fprintln(stdout, "LAN:")
		for _, url := range status.LANURLs {
			fmt.Fprintf(stdout, "  %s\n", url)
		}
	}

	printRemoteOption(stdout, "Tailscale", status.Tailscale)
	fmt.Fprintf(stdout, "Recommended: %s\n", status.Recommended)

	if len(status.RemoteWarnings) > 0 {
		fmt.Fprintln(stdout, "Notes:")
		for _, warning := range status.RemoteWarnings {
			if warning != "" {
				fmt.Fprintf(stdout, "  %s\n", warning)
			}
		}
	}
	return 0
}

func printRemoteOption(stdout io.Writer, name string, option services.RemoteOption) {
	switch {
	case option.URL != "":
		fmt.Fprintf(stdout, "%s: %s\n", name, option.URL)
	case option.Message != "":
		fmt.Fprintf(stdout, "%s: %s\n", name, option.Message)
	default:
		fmt.Fprintf(stdout, "%s: unavailable\n", name)
	}
}

func runStatus(stdout, stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "status failed: %v\n", err)
		return 1
	}

	cfgPath, _ := config.Path()
	fmt.Fprintf(stdout, "Config: %s\n", cfgPath)
	fmt.Fprintf(stdout, "Storage roots: %d\n", len(cfg.StorageRoots))
	var total uint64
	for _, root := range cfg.StorageRoots {
		if root.Enabled {
			total += root.LimitBytes
		}
	}
	fmt.Fprintf(stdout, "Allocated capacity: %s\n", formatSize(total))
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "ParaChute turns this machine into a self-hosted cloud storage node.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  parachute setup [--path <path>] [--limit 100GB]")
	fmt.Fprintln(w, "  parachute storage add <path> --limit 500GB")
	fmt.Fprintln(w, "  parachute storage list")
	fmt.Fprintln(w, "  parachute storage remove <path>")
	fmt.Fprintln(w, "  parachute server start")
	fmt.Fprintln(w, "  parachute remote status")
	fmt.Fprintln(w, "  parachute status")
}

func printStorageHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  parachute storage add <path> --limit 500GB")
	fmt.Fprintln(w, "  parachute storage list")
	fmt.Fprintln(w, "  parachute storage remove <path>")
}

func Main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
