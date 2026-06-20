package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Health int

const (
	HealthClean Health = iota
	HealthAhead
	HealthBehind
	HealthDiverged
	HealthDirty
	HealthDetached
	HealthUnknown
)

func (h Health) String() string {
	switch h {
	case HealthClean:
		return "clean"
	case HealthAhead:
		return "ahead"
	case HealthBehind:
		return "behind"
	case HealthDiverged:
		return "diverged"
	case HealthDirty:
		return "dirty"
	case HealthDetached:
		return "detached"
	default:
		return "unknown"
	}
}

func (h Health) Icon() string {
	switch h {
	case HealthClean:
		return "●"
	case HealthAhead:
		return "↑"
	case HealthBehind:
		return "↓"
	case HealthDiverged:
		return "↕"
	case HealthDirty:
		return "●"
	case HealthDetached:
		return "◆"
	default:
		return "?"
	}
}

type Status struct {
	Modified  int
	Untracked int
	Staged    int
	IsClean   bool
}

func (s Status) Icon() string {
	if s.IsClean {
		return "✓"
	}
	return "M"
}

type Commit struct {
	Message string
	Date    time.Time
}

func (c *Commit) RelativeTime() string {
	if c == nil {
		return ""
	}
	d := time.Since(c.Date)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}

func (c *Commit) IsStale() bool {
	if c == nil {
		return false
	}
	return time.Since(c.Date) > 90*24*time.Hour
}

type Remote struct {
	Name string
	URL  string
}

type Info struct {
	Path       string
	Branch     string
	Status     Status
	Health     Health
	Ahead      int
	Behind     int
	LastCommit *Commit
	Remotes    []Remote
	Detached   bool
	Err        error
}

func (i *Info) HasRemote() bool {
	return len(i.Remotes) > 0
}

func (i *Info) RemoteSummary() string {
	if len(i.Remotes) == 0 {
		return "local"
	}
	var names []string
	seen := make(map[string]bool)
	for _, r := range i.Remotes {
		label := remoteLabel(r.URL)
		if !seen[label] {
			seen[label] = true
			names = append(names, label)
		}
	}
	return strings.Join(names, "+")
}

func remoteLabel(url string) string {
	switch {
	case strings.Contains(url, "github.com"):
		return "github"
	case strings.Contains(url, "codeberg.org"):
		return "codeberg"
	case strings.Contains(url, "gitlab.com"):
		return "gitlab"
	case strings.Contains(url, "bitbucket.org"):
		return "bitbucket"
	case strings.Contains(url, "sr.ht"):
		return "sourcehut"
	case strings.Contains(url, "forgejo"):
		return "forgejo"
	case strings.Contains(url, "gitea.com"):
		return "gitea"
	case strings.Contains(url, "gitea"):
		return "gitea"
	case strings.Contains(url, "notabug.org"):
		return "notabug"
	case strings.Contains(url, "gitea."):
		return "gitea"
	default:
		return "remote"
	}
}

func runGit(path string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", path}, args...)
	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func GetInfo(path string) Info {
	info := Info{Path: path}

	branch, err := getBranch(path)
	if err != nil {
		info.Err = err
		info.Branch = "unknown"
		return info
	}
	info.Branch = branch
	info.Detached = branch == "HEAD"

	status, err := getStatus(path)
	if err != nil {
		info.Err = err
	} else {
		info.Status = status
	}

	commit, err := getLastCommit(path)
	if err == nil {
		info.LastCommit = commit
	}

	ahead, behind, err := getAheadBehind(path)
	if err == nil {
		info.Ahead = ahead
		info.Behind = behind
	}

	remotes, err := getRemotes(path)
	if err == nil {
		info.Remotes = remotes
	}

	info.Health = determineHealth(info)
	return info
}

func getBranch(path string) (string, error) {
	output, err := runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return output, nil
}

func getStatus(path string) (Status, error) {
	output, err := runGit(path, "status", "--porcelain")
	if err != nil {
		return Status{}, err
	}

	if output == "" {
		return Status{IsClean: true}, nil
	}

	var s Status
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		x := line[0]
		y := line[1]

		if x == '?' && y == '?' {
			s.Untracked++
		} else {
			if x != ' ' && x != '?' {
				s.Staged++
			}
			if y != ' ' && y != '?' {
				s.Modified++
			}
			if y == '?' {
				s.Untracked++
			}
		}
	}

	s.IsClean = s.Modified == 0 && s.Untracked == 0 && s.Staged == 0
	return s, nil
}

func getLastCommit(path string) (*Commit, error) {
	output, err := runGit(path, "log", "-1", "--format=%s|%ct")
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(output, "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected log format")
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Commit{
		Message: parts[0],
		Date:    time.Unix(timestamp, 0),
	}, nil
}

func getRemotes(path string) ([]Remote, error) {
	output, err := runGit(path, "remote", "-v")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var remotes []Remote
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		url := parts[1]
		key := name + "|" + url
		if !seen[key] {
			seen[key] = true
			remotes = append(remotes, Remote{Name: name, URL: url})
		}
	}
	return remotes, nil
}

func getAheadBehind(path string) (int, int, error) {
	output, err := runGit(path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(output)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list format")
	}

	ahead, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	behind, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

func determineHealth(info Info) Health {
	if info.Detached {
		return HealthDetached
	}
	if !info.Status.IsClean {
		return HealthDirty
	}
	if info.Ahead > 0 && info.Behind > 0 {
		return HealthDiverged
	}
	if info.Ahead > 0 {
		return HealthAhead
	}
	if info.Behind > 0 {
		return HealthBehind
	}
	return HealthClean
}
