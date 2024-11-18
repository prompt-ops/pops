// FILE: cmd2/common/textinput.go
package common

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInputModel struct {
	textinput textinput.Model
	err       error
}

func NewTextInputModel(placeholder string) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return TextInputModel{
		textinput: ti,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	return m.textinput.View()
}

func (m TextInputModel) Value() string {
	return m.textinput.Value()
}
