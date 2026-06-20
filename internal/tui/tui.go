package tui

import (
	"fmt"
	"strings"
	"time"

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
			Foreground(lipgloss.Color("36")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255"))

	selectedPrefix = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	normalPrefix = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

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
	colRepo      = 28
	colBranch    = 14
	colStatus    = 5
	colHealth    = 12
	colRemote    = 18
	colLastCommit = 16
)

func New(cfg *config.Config, cfgPath string) model {
	return model{
		cfg:     cfg,
		cfgPath: cfgPath,
	}
}

func (m model) Init() tea.Cmd {
	return m.scan()
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

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
		return m, m.scan()

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
	switch msg.String() {

	case "esc":
		m.filtering = false
		m.filterText = ""
		m.applyFilter()

	case "enter":
		m.filtering = false
		m.applyFilter()

	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.applyFilter()
		}

	default:
		if len(msg.String()) == 1 {
			m.filterText += msg.String()
			m.applyFilter()
		}
	}

	return m, nil
}

func (m model) handleAddInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "esc":
		m.adding = false
		m.addText = ""

	case "enter":
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
					return m, m.scan()
				}
			}
		}
		m.adding = false
		m.addText = ""

	case "backspace":
		if len(m.addText) > 0 {
			m.addText = m.addText[:len(m.addText)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.addText += msg.String()
		}
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
		b.WriteString(dimStyle.Render("  Scanning..."))
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errStyle.Render("  Error: " + m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderTable())
	b.WriteString("\n\n")

	b.WriteString(m.renderFooter())

	if m.filtering {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("  / ") + m.filterText)
	}

	if m.adding {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("  add folder: ") + m.addText)
		b.WriteString("  " + dimStyle.Render("(Esc to cancel)"))
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

	var b strings.Builder

	header := m.renderHeaderRow()
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", max(1, min(m.width-2, 120)))))
	b.WriteString("\n")

	for i, repo := range m.filtered {
		line := m.renderRepoRow(repo, i == m.cursor)
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

func (m model) renderRepoRow(repo scanner.Repo, selected bool) string {
	name := padStr(truncate(repo.Name, colRepo-1), colRepo)

	branch := "—"
	statusIcon := "—"
	healthStr := "—"
	remoteStr := "local"
	lastCommit := "—"
	path := truncatePath(repo.Path, 35)

	if repo.GitInfo != nil {
		gi := repo.GitInfo

		if gi.Detached {
			branch = detachedStyle.Render(padStr("(detached)", colBranch-1))
		} else {
			branch = padStr(truncate(gi.Branch, colBranch-1), colBranch)
		}

		if gi.Status.IsClean {
			statusIcon = cleanStyle.Render("✓")
		} else {
			statusIcon = dirtyStyle.Render("M")
		}
		statusIcon = padStr(statusIcon, colStatus)

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
		branch = padStr("—", colBranch)
		statusIcon = padStr("—", colStatus)
		healthStr = padStr("—", colHealth)
		remoteStr = localStyle.Render(padStr("no git", colRemote-1))
		lastCommit = padStr("—", colLastCommit)
	}

	prefix := "  "
	if selected {
		prefix = selectedPrefix.Render("▶ ")
	}

	row := name + branch + statusIcon + healthStr + remoteStr + lastCommit + dimStyle.Render(path)

	if selected {
		return prefix + selectedStyle.Render(row)
	}
	return prefix + normalPrefix.Render("  ") + row
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
	if gi == nil {
		return padStr("—", width)
	}
	summary := gi.RemoteSummary()
	if summary == "local" {
		return localStyle.Render(padStr("◉ local", width-1))
	}
	return cloudStyle.Render(padStr("☁ "+summary, width-1))
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

func truncatePath(p string, n int) string {
	if len(p) <= n {
		return p
	}
	return "…" + p[len(p)-(n-1):]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
