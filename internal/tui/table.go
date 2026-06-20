package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"delta/internal/git"
	"delta/internal/scanner"
)

func (m model) renderHeader() string {
	title := titleStyle.Render("delta")
	subtitle := subtitleStyle.Render("  repo scanner v0.2.0")
	return title + subtitle
}

func (m model) renderTable() string {
	if len(m.filtered) == 0 {
		if len(m.repos) == 0 {
			return dimStyle.Render("  No repos found. Press 'a' to add a scan folder.")
		}
		return dimStyle.Render("  No repos match filter: \"" + m.filterText + "\"")
	}

	pathWidth := m.calcPathWidth()

	var b strings.Builder

	b.WriteString(m.renderHeaderRow())
	b.WriteString("\n")
	dividerWidth := fixedCols + pathWidth + 2
	b.WriteString(dimStyle.Render("  " + strings.Repeat("-", max(1, min(dividerWidth, m.width-2)))))
	b.WriteString("\n")

	visibleStart, visibleEnd := m.getVisibleRange()

	for i := visibleStart; i < visibleEnd; i++ {
		line := m.renderRepoRow(m.filtered[i], i == m.cursor, pathWidth)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderHeaderRow() string {
	cols := []string{
		padStr("REPO", colRepo),
		padStr("BRANCH", colBranch),
		padStr("ST", colStatus),
		padStr("HEALTH", colHealth),
		padStr("REMOTE", colRemote),
		padStr("LAST COMMIT", colLastCommit),
		"PATH",
	}
	return "  " + headerStyle.Render(strings.Join(cols, ""))
}

func (m model) renderRepoRow(repo scanner.Repo, selected bool, pathWidth int) string {
	name := padStr(truncate(m.displayName(repo), colRepo-1), colRepo)
	branch := padStr("-", colBranch)
	statusIcon := padStr("-", colStatus)
	healthStr := padStr("-", colHealth)
	remoteStr := padStr("-", colRemote)
	lastCommit := padStr("-", colLastCommit)
	path := shortenPath(repo.Path, pathWidth)

	if repo.GitInfo != nil {
		gi := repo.GitInfo

		if gi.Detached {
			branch = padStr(detachedStyle.Render("(detached)"), colBranch)
		} else {
			branch = padStr(truncate(gi.Branch, colBranch-1), colBranch)
		}

		if gi.Status.IsClean {
			statusIcon = padStr(cleanStyle.Render("v"), colStatus)
		} else {
			statusIcon = padStr(dirtyStyle.Render("M"), colStatus)
		}

		healthStr = renderHealth(gi.Health, colHealth)
		remoteStr = renderRemote(gi, colRemote)

		if gi.LastCommit != nil {
			lastCommit = gi.LastCommit.RelativeTime()
			if gi.LastCommit.IsStale() {
				lastCommit = staleStyle.Render("! " + lastCommit)
			}
		}
		lastCommit = padStr(lastCommit, colLastCommit)
	} else {
		remoteStr = localStyle.Render(padStr("no git", colRemote-1))
	}

	row := name + branch + statusIcon + healthStr + remoteStr + lastCommit + dimStyle.Render(path)

	if selected {
		return selectedStyle.Render(" > " + row)
	}
	return "   " + row
}

func (m model) displayName(repo scanner.Repo) string {
	if !m.hasDuplicateName(repo.Name) {
		return repo.Name
	}
	parent := filepath.Base(filepath.Dir(repo.Path))
	return repo.Name + " (" + parent + ")"
}

func (m model) hasDuplicateName(name string) bool {
	count := 0
	for _, r := range m.repos {
		if r.Name == name {
			count++
			if count > 1 {
				return true
			}
		}
	}
	return false
}

func shortenPath(p string, maxLen int) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(p, home) {
		p = "~" + strings.TrimPrefix(p, home)
	}

	if len(p) <= maxLen {
		return p
	}

	parts := strings.Split(p, string(filepath.Separator))
	if len(parts) <= 2 {
		return p
	}

	for len(parts) > 2 {
		mid := len(parts) / 2
		parts = append(parts[:mid], parts[mid+1:]...)
		result := strings.Join(parts, string(filepath.Separator))
		if len(result) <= maxLen {
			return result
		}
	}

	if len(p) > maxLen {
		return "..." + p[len(p)-(maxLen-3):]
	}
	return p
}

func renderHealth(h git.Health, width int) string {
	var icon, label string
	switch h {
	case git.HealthClean:
		icon = cleanStyle.Render("o")
		label = "clean"
	case git.HealthAhead:
		icon = aheadStyle.Render("^")
		label = "ahead"
	case git.HealthBehind:
		icon = behindStyle.Render("v")
		label = "behind"
	case git.HealthDiverged:
		icon = dirtyStyle.Render("/\\")
		label = "diverged"
	case git.HealthDirty:
		icon = dirtyStyle.Render("o")
		label = "dirty"
	case git.HealthDetached:
		icon = detachedStyle.Render("<>")
		label = "detached"
	default:
		icon = "?"
		label = "unknown"
	}
	return padStr(icon+" "+label, width)
}

func renderRemote(gi *git.Info, width int) string {
	if gi == nil || len(gi.Remotes) == 0 {
		return localStyle.Render(padStr("local", width-1))
	}
	summary := gi.RemoteSummary()
	return cloudStyle.Render(padStr(summary, width-1))
}

func (m model) renderFooter() string {
	parts := []string{itoa(len(m.filtered)) + " repos"}
	if m.filterText != "" && !m.filtering {
		parts = append(parts, "filtered from "+itoa(len(m.repos)))
	}
	if !m.lastScan.IsZero() {
		parts = append(parts, "scan: "+m.scanTime.Round(1000000).String())
	}

	var b strings.Builder
	b.WriteString(footerStyle.Render("  " + strings.Join(parts, "  -  ")))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  [up/down] navigate  [r] refresh  [/] filter  [a] add folder  [q] quit"))
	return b.String()
}

func (m model) getVisibleRange() (int, int) {
	total := len(m.filtered)
	if total == 0 {
		return 0, 0
	}

	maxRows := m.height - 8
	if maxRows < 5 {
		maxRows = 5
	}
	if maxRows > total {
		maxRows = total
	}

	if m.cursor >= total {
		m.cursor = total - 1
	}

	if m.cursor < maxRows {
		return 0, maxRows
	}

	start := m.cursor - maxRows + 1
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > total {
		end = total
	}
	return start, end
}

func (m model) calcPathWidth() int {
	used := fixedCols + 5
	available := m.width - used
	if available < 20 {
		return 20
	}
	if available > 60 {
		return 60
	}
	return available
}

func padStr(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func truncate(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	return s[:n-1] + "..."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
