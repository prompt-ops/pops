package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type enterNameModel struct {
	textInput textinput.Model
	err       error
	quitting  bool
	done      bool
}

func initialEnterNameModel() enterNameModel {
	ti := textinput.New()
	ti.Placeholder = "My Cloud Connection"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return enterNameModel{
		textInput: ti,
		err:       nil,
		quitting:  false,
		done:      false,
	}
}

func (m enterNameModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m enterNameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m enterNameModel) View() string {
	if m.err != nil {
		return fmt.Sprintf(
			"Error: %v\n\nPress any key to try again...",
			m.err,
		)
	}
	return fmt.Sprintf(
		"Enter a connection name\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func (m enterNameModel) Quitting() bool {
	return m.quitting
}

func (m enterNameModel) Value() string {
	return m.textInput.Value()
}

func (m enterNameModel) Done() bool {
	return m.done
}

func NewEnterNameModel() enterNameModel {
	return initialEnterNameModel()
}
