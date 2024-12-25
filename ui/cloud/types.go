package cloud

import "github.com/charmbracelet/lipgloss"

type step int

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
