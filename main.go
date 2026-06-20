package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"delta/internal/config"
	"delta/internal/scanner"
	"delta/internal/tui"
)

func main() {
	configPath := flag.String("c", "", "path to config file (default: ~/.config/delta/config.json)")
	scanOnly := flag.Bool("scan", false, "scan folders and print results as JSON, then exit")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	resolvedPath := *configPath
	if resolvedPath == "" {
		resolvedPath, _ = config.DefaultPath()
	}

	if *scanOnly {
		repos, err := scanner.ScanFolders(cfg.ScanFolders)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error scanning: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("found %d repos\n", len(repos))
		for _, r := range repos {
			fmt.Printf("  %s (%s)\n", r.Name, r.Path)
		}
		return
	}

	p := tea.NewProgram(tui.New(cfg, resolvedPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
