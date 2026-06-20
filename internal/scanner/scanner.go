package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"delta/internal/git"
)

type Repo struct {
	Path     string
	Name     string
	IsGit    bool
	GitInfo  *git.Info
}

func ScanFolders(folders []string) ([]Repo, error) {
	var repos []Repo

	for _, folder := range folders {
		folder = strings.Trim(folder, "\"")
		folder = filepath.Clean(folder)

		info, err := os.Stat(folder)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("could not access %s: %w", folder, err)
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				return nil
			}

			base := filepath.Base(path)
			if base == ".git" {
				parent := filepath.Dir(path)
				repos = append(repos, Repo{
					Path:  parent,
					Name:  filepath.Base(parent),
					IsGit: true,
				})
				return filepath.SkipDir
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking %s: %w", folder, err)
		}
	}

	repos = dedup(repos)
	return repos, nil
}

func ScanFoldersWithGitInfo(folders []string) ([]Repo, error) {
	repos, err := ScanFolders(folders)
	if err != nil {
		return nil, err
	}

	for i := range repos {
		if repos[i].IsGit {
			info := git.GetInfo(repos[i].Path)
			repos[i].GitInfo = &info
		}
	}

	return repos, nil
}

func dedup(repos []Repo) []Repo {
	seen := make(map[string]bool)
	result := make([]Repo, 0, len(repos))
	for _, r := range repos {
		if !seen[r.Path] {
			seen[r.Path] = true
			result = append(result, r)
		}
	}
	return result
}
