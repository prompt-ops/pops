package common

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

type spinnerMsg struct{}

type spinnerModel struct {
	spinner spinner.Model
	msg     string
	err     error
}

func newSpinnerModel(msg string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return spinnerModel{
		spinner: s,
		msg:     msg,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		sp, cmd := m.spinner.Update(msg)
		m.spinner = sp
		return m, cmd
	case spinnerMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m spinnerModel) View() string {
	return fmt.Sprintf("%s %s", m.spinner.View(), m.msg)
}

// RunWithSpinner runs a Bubble Tea program with a spinner and executes the provided function.
func RunWithSpinner(msg string, fn func() error) error {
	model := newSpinnerModel(msg)
	p := tea.NewProgram(model)

	// Run the spinner in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second)

		err := fn()
		if err != nil {
			color.Red("Error: %v", err)
		} else {
			color.Green("Success")
		}

		p.Send(spinnerMsg{})
	}()

	// Start the spinner program
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
