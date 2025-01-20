package shell

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Padding(0, 1)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Padding(0, 1)

	confirmationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("178")).
				Padding(0, 1)

	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), true)

		// History related styles
	historyContainerStyle = lipgloss.NewStyle().
				Width(72).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				Margin(1, 0)

	historyLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))
)
