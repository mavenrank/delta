package tui

import "github.com/charmbracelet/lipgloss"

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
