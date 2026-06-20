package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"delta/internal/config"
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
			Foreground(lipgloss.Color("63"))

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

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	inputPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))
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
			repos: repos,
			err:   err,
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

	b.WriteString(titleStyle.Render("delta") + dimStyle.Render("  v0.2.0"))
	b.WriteString("\n\n")

	if m.scanning {
		b.WriteString(dimStyle.Render("Scanning..."))
		b.WriteString("\n")
	}

	if m.err != "" {
		b.WriteString(errStyle.Render("Error: "+m.err))
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderTable())
	b.WriteString("\n\n")

	b.WriteString(m.renderFooter())

	if m.filtering {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("/ ") + m.filterText)
	}

	if m.adding {
		b.WriteString("\n\n")
		b.WriteString(inputPromptStyle.Render("add folder: ") + m.addText)
	}

	return b.String()
}

func (m model) renderTable() string {
	if len(m.filtered) == 0 {
		if len(m.repos) == 0 {
			return dimStyle.Render("No repos found. Press 'a' to add a scan folder.")
		}
		return dimStyle.Render("No repos match filter: \"" + m.filterText + "\"")
	}

	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("%-25s %-12s %-4s %-8s %-20s %s",
		"REPO", "BRANCH", "ST", "HEALTH", "LAST COMMIT", "PATH")))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(m.width, 120)))
	b.WriteString("\n")

	for i, repo := range m.filtered {
		line := m.renderRepoLine(repo)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderRepoLine(repo scanner.Repo) string {
	name := truncate(repo.Name, 25)
	branch := "—"
	statusIcon := "—"
	healthStr := "—"
	healthIcon := "?"
	lastCommit := "—"
	path := truncatePath(repo.Path, 30)

	if repo.GitInfo != nil {
		gi := repo.GitInfo

		branch = truncate(gi.Branch, 12)
		if gi.Detached {
			branch = detachedStyle.Render(branch)
		}

		statusIcon = gi.Status.Icon()
		if gi.Status.IsClean {
			statusIcon = cleanStyle.Render(statusIcon)
		} else {
			statusIcon = dirtyStyle.Render(statusIcon)
		}

		healthIcon = gi.Health.Icon()
		healthStr = gi.Health.String()
		switch gi.Health {
		case 1:
			healthStr = aheadStyle.Render(healthStr)
		case 2:
			healthStr = behindStyle.Render(healthStr)
		case 3:
			healthStr = dirtyStyle.Render(healthStr)
		case 5:
			healthStr = detachedStyle.Render(healthStr)
		default:
			healthStr = cleanStyle.Render(healthStr)
		}
		healthStr = fmt.Sprintf("%s %s", healthIcon, healthStr)

		if gi.LastCommit != nil {
			lastCommit = gi.LastCommit.RelativeTime()
			if gi.LastCommit.IsStale() {
				lastCommit = staleStyle.Render("⚠ " + lastCommit)
			}
		}
	}

	return fmt.Sprintf("%-25s %-12s %-4s %-14s %-20s %s",
		name, branch, statusIcon, healthStr, truncate(lastCommit, 20), path)
}

func (m model) renderFooter() string {
	var b strings.Builder

	b.WriteString(footerStyle.Render(fmt.Sprintf("%d repos", len(m.filtered))))
	if m.filterText != "" && !m.filtering {
		b.WriteString(footerStyle.Render(fmt.Sprintf(" (filtered from %d)", len(m.repos))))
	}
	if !m.lastScan.IsZero() {
		b.WriteString(footerStyle.Render(fmt.Sprintf("  ·  scan: %s", m.scanTime.Round(time.Millisecond))))
	}
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("[↑↓] navigate  [r] refresh  [/] filter  [a] add folder  [q] quit"))

	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
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
