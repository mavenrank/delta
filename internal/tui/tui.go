package tui

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"delta/internal/config"
	"delta/internal/git"
	"delta/internal/scanner"
)

type scanMsg struct {
	repos    []scanner.Repo
	err      error
	scanTime time.Duration
}

type spinnerMsg struct{}

type model struct {
	repos      []scanner.Repo
	filtered   []scanner.Repo
	cursor     int
	cfg        *config.Config
	cfgPath    string
	filtering  bool
	filterText string
	adding     bool
	addText    string
	width      int
	height     int
	err        string
	scanning   bool
	scanTime   time.Duration
	lastScan   time.Time
	quit       bool
	spinnerIdx int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("36"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	aheadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	behindStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("215"))

	staleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("227"))

	detachedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("201"))

	localStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("130"))

	cloudStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	inputPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))
)

const (
	colRepo       = 28
	colBranch     = 14
	colStatus     = 5
	colHealth     = 12
	colRemote     = 18
	colLastCommit = 16
	fixedCols     = colRepo + colBranch + colStatus + colHealth + colRemote + colLastCommit
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func New(cfg *config.Config, cfgPath string) model {
	return model{
		cfg:      cfg,
		cfgPath:  cfgPath,
		scanning: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.scan(), m.spinnerTick())
}

func (m model) scan() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		repos, err := scanner.ScanFoldersWithGitInfo(m.cfg.ScanFolders)
		duration := time.Since(start)
		return scanMsg{
			repos:    repos,
			err:      err,
			scanTime: duration,
		}
	}
}

func (m model) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case spinnerMsg:
		if m.scanning {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, m.spinnerTick()
		}
		return m, nil

	case scanMsg:
		m.scanning = false
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.err = ""
			m.repos = msg.repos
			m.scanTime = msg.scanTime
			m.lastScan = time.Now()
			m.applyFilter()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.adding {
			return m.handleAddInput(msg)
		}
		if m.filtering {
			return m.handleFilterInput(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "q", "ctrl+c":
		m.quit = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case "r":
		m.scanning = true
		m.err = ""
		m.repos = nil
		m.filtered = nil
		return m, tea.Batch(m.scan(), m.spinnerTick())

	case "/":
		m.filtering = true
		m.filterText = ""

	case "a":
		m.adding = true
		m.addText = ""
	}

	return m, nil
}

func (m model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case tea.KeyEscape:
		m.filtering = false
		m.filterText = ""
		m.applyFilter()
		return m, nil

	case tea.KeyEnter:
		m.filtering = false
		m.applyFilter()
		return m, nil

	case tea.KeyBackspace:
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.applyFilter()
		}
		return m, nil

	case tea.KeyCtrlC:
		m.quit = true
		return m, tea.Quit

	case tea.KeyRunes:
		runes := msg.Runes
		for _, r := range runes {
			if unicode.IsPrint(r) {
				m.filterText += string(r)
			}
		}
		m.applyFilter()
		return m, nil
	}

	return m, nil
}

func (m model) handleAddInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case tea.KeyEscape:
		m.adding = false
		m.addText = ""
		return m, nil

	case tea.KeyEnter:
		path := strings.Trim(m.addText, "\"")
		if path != "" {
			err := m.cfg.AddFolder(path)
			if err != nil {
				m.err = err.Error()
			} else {
				err = m.cfg.Save(m.cfgPath)
				if err != nil {
					m.err = err.Error()
				} else {
					m.err = ""
					m.scanning = true
					m.adding = false
					m.addText = ""
					return m, tea.Batch(m.scan(), m.spinnerTick())
				}
			}
		}
		m.adding = false
		m.addText = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.addText) > 0 {
			m.addText = m.addText[:len(m.addText)-1]
		}
		return m, nil

	case tea.KeyCtrlC:
		m.quit = true
		return m, tea.Quit

	case tea.KeyRunes:
		runes := msg.Runes
		for _, r := range runes {
			if unicode.IsPrint(r) {
				m.addText += string(r)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m *model) applyFilter() {
	if m.filterText == "" {
		m.filtered = m.repos
	} else {
		filtered := make([]scanner.Repo, 0)
		for _, r := range m.repos {
			if strings.Contains(strings.ToLower(r.Name), strings.ToLower(m.filterText)) {
				filtered = append(filtered, r)
			}
		}
		m.filtered = filtered
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
	if m.quit {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	if m.scanning {
		spinner := spinnerFrames[m.spinnerIdx]
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %s scanning folders...", spinner)))
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errStyle.Render("  Error: " + m.err))
		b.WriteString("\n\n")
	}

	if m.scanning && len(m.repos) == 0 {
		b.WriteString(dimStyle.Render("  Looking for repositories..."))
		b.WriteString("\n\n")
	} else {
		b.WriteString(m.renderTable())
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderFooter())

	if m.filtering {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("  / ") + m.filterText)
		b.WriteString(dimStyle.Render("  (Enter to apply, Esc to cancel)"))
	}

	if m.adding {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("  add folder: ") + m.addText)
		b.WriteString(dimStyle.Render("  (Enter to save, Esc to cancel)"))
	}

	return b.String()
}

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

	header := m.renderHeaderRow()
	b.WriteString(header)
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
		b.WriteString(dimStyle.Render(fmt.Sprintf("  ... %d more repos below", len(m.filtered)-visibleEnd)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) getVisibleRange() (int, int) {
	maxRows := m.height - 10
	if maxRows < 5 {
		maxRows = 5
	}
	if maxRows > len(m.filtered) {
		maxRows = len(m.filtered)
	}

	if m.cursor < maxRows/2 {
		return 0, maxRows
	}

	if m.cursor > len(m.filtered)-maxRows/2 {
		return max(0, len(m.filtered)-maxRows), len(m.filtered)
	}

	start := m.cursor - maxRows/2
	return start, start + maxRows
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
	var b strings.Builder

	parts := []string{fmt.Sprintf("%d repos", len(m.filtered))}
	if m.filterText != "" && !m.filtering {
		parts = append(parts, fmt.Sprintf("filtered from %d", len(m.repos)))
	}
	if !m.lastScan.IsZero() {
		parts = append(parts, fmt.Sprintf("scan: %s", m.scanTime.Round(time.Millisecond)))
	}

	b.WriteString(footerStyle.Render("  " + strings.Join(parts, "  ·  ")))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  [↑↓] navigate  [r] refresh  [/] filter  [a] add folder  [q] quit"))

	return b.String()
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
