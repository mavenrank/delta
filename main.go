package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"delta/internal/config"
	"delta/internal/scanner"
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

	repos, nonGit, err := scanner.ScanFolders(cfg.ScanFolders)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning folders: %v\n", err)
		os.Exit(1)
	}

	if *scanOnly {
		fmt.Printf("{\"repos\": %d, \"non_git\": %d}\n", len(repos), len(nonGit))
		return
	}

	fmt.Printf("delta v%s\n", version())
	fmt.Printf("found %d repos, %d non-git folders\n", len(repos), len(nonGit))
	fmt.Println("TUI not yet implemented (planned for v0.2.0)")
	for _, r := range repos {
		fmt.Printf("  %s (%s)\n", filepath.Base(r), r)
	}
}

func version() string {
	return "0.1.0"
}
