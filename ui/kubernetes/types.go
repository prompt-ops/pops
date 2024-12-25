package kubernetes

import "github.com/charmbracelet/lipgloss"

type step int

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
