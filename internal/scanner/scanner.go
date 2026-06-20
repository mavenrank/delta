package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"delta/internal/git"
)

type Repo struct {
	Path    string
	Name    string
	IsGit   bool
	GitInfo *git.Info
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

		walkAndCollect(folder, &repos, false)
	}

	repos = dedup(repos)
	return repos, nil
}

func walkAndCollect(root string, repos *[]Repo, insideGit bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	hasGit := false
	for _, entry := range entries {
		if entry.Name() == ".git" {
			hasGit = true
			break
		}
	}

	if hasGit {
		*repos = append(*repos, Repo{
			Path:  root,
			Name:  filepath.Base(root),
			IsGit: true,
		})
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if shouldSkipDir(entry.Name()) {
				continue
			}
			walkAndCollect(filepath.Join(root, entry.Name()), repos, true)
		}
		return
	}

	if !insideGit && isNonGitCodeFolder(root) && !hasSubRepo(root) {
		*repos = append(*repos, Repo{
			Path:  root,
			Name:  filepath.Base(root),
			IsGit: false,
		})
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if shouldSkipDir(name) {
			continue
		}
		walkAndCollect(filepath.Join(root, name), repos, insideGit)
	}
}

func shouldSkipDir(name string) bool {
	skip := []string{
		"node_modules", ".venv", "venv", "__pycache__",
		".next", ".nuxt", "dist", "build", ".cache",
		"target", ".gradle", ".idea", ".vscode",
		"env", ".env", "vendor", "Pods", "bin", "obj",
		".git", ".hg", ".svn", "site-packages",
		".langchain-venv", "ephemeral", "Runner",
		"RunnerTests", "Flutter",
	}
	for _, s := range skip {
		if name == s {
			return true
		}
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	return false
}

func hasSubRepo(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if shouldSkipDir(entry.Name()) {
			continue
		}
		sub := filepath.Join(path, entry.Name())
		if dirHasGit(sub) {
			return true
		}
	}
	return false
}

func dirHasGit(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

func isNonGitCodeFolder(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			return false
		}
		if isProjectMarker(name) {
			return true
		}
		if !entry.IsDir() && isCodeFile(name) {
			return true
		}
	}
	return false
}

func isProjectMarker(name string) bool {
	markers := []string{
		"package.json", "go.mod", "Cargo.toml",
		"requirements.txt", "pyproject.toml",
		"Makefile", "CMakeLists.txt", "pom.xml",
		"build.gradle", "package-lock.json",
		"yarn.lock", "pnpm-lock.yaml",
		"tsconfig.json", "composer.json",
	}
	for _, m := range markers {
		if name == m {
			return true
		}
	}
	return false
}

func isCodeFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	codeExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".c", ".cpp", ".h", ".hpp", ".rs", ".java",
		".rb", ".php", ".cs", ".swift", ".kt",
	}
	for _, e := range codeExts {
		if ext == e {
			return true
		}
	}
	return false
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
