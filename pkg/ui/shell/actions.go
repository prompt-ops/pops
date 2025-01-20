package shell

import tea "github.com/charmbracelet/bubbletea"

func (m shellModel) runInitialChecks() tea.Msg {
	err := m.popsConnection.CheckAuthentication()
	if err != nil {
		return errMsg{err}
	}

	err = m.popsConnection.SetContext()
	if err != nil {
		return errMsg{err}
	}

	return checkPassedMsg{}
}

func (m shellModel) generateCommand(prompt string) tea.Cmd {
	return func() tea.Msg {
		cmd, err := m.popsConnection.GetCommand(prompt)
		if err != nil {
			return errMsg{err}
		}

		return commandMsg{
			command: cmd,
		}
	}
}

func (m shellModel) runCommand(command string) tea.Cmd {
	return func() tea.Msg {
		out, err := m.popsConnection.ExecuteCommand(command)
		if err != nil {
			return errMsg{err}
		}

		outStr, err := m.popsConnection.FormatResultAsTable(out)
		if err != nil {
			return errMsg{err}
		}

		return outputMsg{
			output: outStr,
		}
	}
}

func (m shellModel) generateAnswer(prompt string) tea.Cmd {
	return func() tea.Msg {
		answer, err := m.popsConnection.GetAnswer(prompt)
		if err != nil {
			return errMsg{err}
		}

		return answerMsg{
			answer,
		}
	}
}
