package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"delta/internal/config"
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

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case tea.MouseButtonWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}
		}
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

func (m model) View() string {
	if m.quit {
		return ""
	}

	footer := m.renderFooter()

	inputLine := ""
	if m.filtering {
		inputLine = inputPromptStyle.Render("  / ") + m.filterText +
			dimStyle.Render("  (Enter to apply, Esc to cancel)")
	}
	if m.adding {
		inputLine = inputPromptStyle.Render("  add folder: ") + m.addText +
			dimStyle.Render("  (Enter to save, Esc to cancel)")
	}

	header := m.renderHeader()

	var content string
	if m.scanning {
		spinner := spinnerFrames[m.spinnerIdx]
		content = dimStyle.Render(fmt.Sprintf("  %s scanning folders...", spinner))
	} else if m.err != "" {
		content = errStyle.Render("  Error: " + m.err)
	} else {
		content = m.renderTable()
	}

	lines := []string{header, "", content}

	availableHeight := m.height - 4
	if inputLine != "" {
		availableHeight--
	}
	if availableHeight < 1 {
		availableHeight = 1
	}

	usedLines := 0
	for _, l := range lines {
		usedLines += max(1, strings.Count(l, "\n")+1)
	}

	if usedLines < availableHeight {
		lines = append(lines, strings.Repeat("\n", availableHeight-usedLines))
	}

	lines = append(lines, "", footer)

	if inputLine != "" {
		lines = append(lines, "", inputLine)
	}

	return strings.Join(lines, "\n")
}
