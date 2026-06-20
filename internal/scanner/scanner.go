package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScanFolders walks each folder recursively, finding Git repos and non-git code folders.
// Returns: (gitRepoPaths, nonGitFolderPaths, error)
func ScanFolders(folders []string) ([]string, []string, error) {
	var repos []string
	var nonGit []string

	for _, folder := range folders {
		folder = strings.Trim(folder, "\"")
		folder = filepath.Clean(folder)

		info, err := os.Stat(folder)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, fmt.Errorf("could not access %s: %w", folder, err)
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
				repos = append(repos, parent)
				return filepath.SkipDir
			}

			if isNonGitCodeFolder(path, info) {
				nonGit = append(nonGit, path)
			}

			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("error walking %s: %w", folder, err)
		}
	}

	return dedup(repos), dedup(nonGit), nil
}

// isNonGitCodeFolder checks if a directory looks like a code project but has no .git.
func isNonGitCodeFolder(path string, info os.FileInfo) bool {
	if info == nil || !info.IsDir() {
		return false
	}

	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return false
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	hasCodeFiles := false
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			return false
		}
		if isCodeFile(name) || isCodeDir(name) {
			hasCodeFiles = true
		}
	}

	return hasCodeFiles
}

func isCodeFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	codeExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".c", ".cpp", ".h", ".hpp", ".rs", ".java",
		".rb", ".php", ".cs", ".swift", ".kt",
		".json", ".yaml", ".yml", ".toml",
	}
	for _, e := range codeExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isCodeDir(name string) bool {
	codeDirs := []string{
		"src", "lib", "bin", "cmd", "internal",
		"pkg", "tests", "test", "vendor",
	}
	for _, d := range codeDirs {
		if name == d {
			return true
		}
	}
	return false
}

func dedup(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
