package shell

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Padding(0, 1)

	commandConfirmationTitleStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Padding(0, 1)

	commandConfirmationContentStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")).
					Padding(0, 1)

	commandConfirmationResponseStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("10")).
						Padding(0, 1)

	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Padding(0, 1)

		// History related styles
	historyContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1).
				Margin(1, 0)

	historyLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	historyCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))
)
