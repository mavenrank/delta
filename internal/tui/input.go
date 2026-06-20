package tui

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"delta/internal/scanner"
)

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
		for _, r := range msg.Runes {
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
		for _, r := range msg.Runes {
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
		m.filtered = filterRepos(m.repos, m.filterText)
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func filterRepos(repos []scanner.Repo, text string) []scanner.Repo {
	lower := strings.ToLower(text)
	result := make([]scanner.Repo, 0)
	for _, r := range repos {
		if strings.Contains(strings.ToLower(r.Name), lower) {
			result = append(result, r)
		}
	}
	return result
}
