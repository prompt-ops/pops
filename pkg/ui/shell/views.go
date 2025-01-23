package shell

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m shellModel) renderFooter(text string) string {
	return footerStyle.Render(text)
}

func (m shellModel) viewInitialChecks() string {
	if m.checkPassed {
		return outputStyle.Render("✅ Authentication passed!\n\n")
	}
	return fmt.Sprintf(
		"%s %s",
		m.spinner.View(),
		titleStyle.Render("Checking Authentication..."),
	)
}

func (m shellModel) viewEnterPrompt() string {
	var title string
	var modeStr string

	if m.mode == modeCommand {
		title = "🤖 Request a command/query"
		modeStr = "command/query"
	} else {
		title = "💡 Ask a question"
		modeStr = "answer"
	}

	footer := m.renderFooter("Use ←/→ to switch between modes (currently " + modeStr + "). Press Enter when ready.\n\nPress F1 to show context.")

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		promptStyle.Render(m.promptInput.View()),
		footer,
	)
}

func (m shellModel) viewShowContext() string {
	footer := m.renderFooter("Press F1 to return to prompt.")

	return fmt.Sprintf(
		"%s\n\n%s",
		titleStyle.Render("ℹ️ Current Context"),
		outputStyle.Render(m.output),
	) + "\n\n" + footer
}

func (m shellModel) viewGenerateCommand() string {
	return titleStyle.Render("🤖 Generating command...")
}

func (m shellModel) viewGetAnswer() string {
	return titleStyle.Render("🤔 Getting your answer...")
}

func (m shellModel) viewConfirmRun() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		commandConfirmationTitleStyle.Render("🚀 Would you like to run the following command? (Y/n)"),
		commandConfirmationContentStyle.Render("🐳 "+m.command),
		commandConfirmationResponseStyle.Render(m.confirmInput.View()),
	)
}

func (m shellModel) viewRunCommand() string {
	return titleStyle.Render("🏃 Running command...")
}

func (m shellModel) viewDone() string {
	width := m.calculateShareViewWidth()

	outStyle := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width)

	var content string
	if m.err != nil {
		content = fmt.Sprintf("%v\n", m.err)
		content = errorStyle.Render(content)
	} else {
		content = fmt.Sprintf("%s\n", m.output)
		content = outputStyle.Render(content)
	}

	content = outStyle.Render(content)
	footer := m.renderFooter("Press 'q' or 'esc' or Ctrl+C to quit, or enter a new prompt.")
	return lipgloss.JoinVertical(lipgloss.Top, content, footer)
}

func (m shellModel) viewHistory() string {
	if len(m.history) == 0 {
		return ""
	}

	var entries []string
	for _, h := range m.history {
		var promptLine string
		var modeLine string
		if h.mode == "Command" {
			promptLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Prompt: "),
				promptStyle.Render(h.prompt),
			)

			modeLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Command: "),
				historyCommandStyle.Render(h.cmd),
			)
		} else {
			promptLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Question: "),
				promptStyle.Render(h.prompt),
			)

			modeLine = lipgloss.JoinHorizontal(
				lipgloss.Top,
				historyLabelStyle.Render("Answer: "),
			)
		}

		outputLine := lipgloss.JoinHorizontal(
			lipgloss.Top,
			outputStyle.Render(h.output),
		)

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			promptLine,
			modeLine,
			outputLine,
		)

		content = lipgloss.JoinVertical(lipgloss.Left, content)

		// Adding minus 10 to the history box width.
		// FIXME: Right border is not visible.
		width := m.calculateShareViewWidth()
		boxed := historyContainerStyle.
			Width(width).
			MaxWidth(width).
			Render(content)
		entries = append(entries, boxed)
	}

	return lipgloss.JoinVertical(lipgloss.Left, entries...)
}

func (m shellModel) calculateShareViewWidth() int {
	maxWidth := m.windowWidth - 2
	if maxWidth < 20 {
		maxWidth = 20
	}
	return maxWidth
}
