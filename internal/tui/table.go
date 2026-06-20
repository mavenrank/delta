package tui

import (
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
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", max(1, min(dividerWidth, m.width-2)))))
	b.WriteString("\n")

	visibleStart, visibleEnd := m.getVisibleRange()

	for i := visibleStart; i < visibleEnd; i++ {
		line := m.renderRepoRow(m.filtered[i], i == m.cursor, pathWidth)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if visibleEnd < len(m.filtered) {
		b.WriteString(dimStyle.Render("  ... " + itoa(len(m.filtered)-visibleEnd) + " more repos below"))
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
	name := padStr(truncate(repo.Name, colRepo-1), colRepo)
	branch := padStr("—", colBranch)
	statusIcon := padStr("—", colStatus)
	healthStr := padStr("—", colHealth)
	remoteStr := padStr("—", colRemote)
	lastCommit := padStr("—", colLastCommit)
	path := truncatePathMiddle(repo.Path, pathWidth)

	if repo.GitInfo != nil {
		gi := repo.GitInfo

		if gi.Detached {
			branch = padStr(detachedStyle.Render("(detached)"), colBranch)
		} else {
			branch = padStr(truncate(gi.Branch, colBranch-1), colBranch)
		}

		if gi.Status.IsClean {
			statusIcon = padStr(cleanStyle.Render("✓"), colStatus)
		} else {
			statusIcon = padStr(dirtyStyle.Render("M"), colStatus)
		}

		healthStr = renderHealth(gi.Health, colHealth)
		remoteStr = renderRemote(gi, colRemote)

		if gi.LastCommit != nil {
			lastCommit = gi.LastCommit.RelativeTime()
			if gi.LastCommit.IsStale() {
				lastCommit = staleStyle.Render("⚠ " + lastCommit)
			}
		}
		lastCommit = padStr(lastCommit, colLastCommit)
	} else {
		remoteStr = localStyle.Render(padStr("no git", colRemote-1))
	}

	row := name + branch + statusIcon + healthStr + remoteStr + lastCommit + dimStyle.Render(path)

	if selected {
		return "▶ " + selectedStyle.Render(" " + row)
	}
	return "  " + row
}

func renderHealth(h git.Health, width int) string {
	var icon, label string
	switch h {
	case git.HealthClean:
		icon = cleanStyle.Render("●")
		label = "clean"
	case git.HealthAhead:
		icon = aheadStyle.Render("↑")
		label = "ahead"
	case git.HealthBehind:
		icon = behindStyle.Render("↓")
		label = "behind"
	case git.HealthDiverged:
		icon = dirtyStyle.Render("↕")
		label = "diverged"
	case git.HealthDirty:
		icon = dirtyStyle.Render("●")
		label = "dirty"
	case git.HealthDetached:
		icon = detachedStyle.Render("◆")
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
	b.WriteString(footerStyle.Render("  " + strings.Join(parts, "  ·  ")))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  [↑↓] navigate  [r] refresh  [/] filter  [a] add folder  [q] quit"))
	return b.String()
}

func (m model) getVisibleRange() (int, int) {
	total := len(m.filtered)
	if total == 0 {
		return 0, 0
	}

	maxRows := m.height - 10
	if maxRows < 5 {
		maxRows = 5
	}
	if maxRows > total {
		maxRows = total
	}

	half := maxRows / 2

	if m.cursor < half {
		return 0, maxRows
	}

	if m.cursor > total-half {
		start := total - maxRows
		if start < 0 {
			start = 0
		}
		return start, total
	}

	start := m.cursor - half
	end := start + maxRows
	if end > total {
		end = total
	}
	return start, end
}

func (m model) calcPathWidth() int {
	used := fixedCols + 4
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
	return s[:n-1] + "…"
}

func truncatePathMiddle(p string, n int) string {
	if len(p) <= n {
		return p
	}
	if n < 10 {
		return "…" + p[len(p)-(n-1):]
	}
	keepStart := n / 3
	keepEnd := n - keepStart - 1
	return p[:keepStart] + "…" + p[len(p)-keepEnd:]
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
