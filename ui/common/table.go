package common

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type tableModel struct {
	table    table.Model
	selected string

	// onSelect is an optional function that will be called
	// when a row is selected if specified.
	onSelect func(string) tea.Msg
}

func NewTableModel(table table.Model, onSelect func(string) tea.Msg) *tableModel {
	return &tableModel{
		table:    table,
		onSelect: onSelect,
	}
}

func (m *tableModel) Init() tea.Cmd {
	return nil
}

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedRow := m.table.SelectedRow()
			if selectedRow == nil {
				// No selection made
				return m, tea.Quit
			}
			m.selected = selectedRow[0]

			// If onSelect is specified, call it.
			if m.onSelect != nil {
				return m, func() tea.Msg {
					return m.onSelect(m.selected)
				}
			}

			return m, tea.Quit
		}
	}

	m.table, _ = m.table.Update(msg)
	return m, nil
}

func (m *tableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func (m *tableModel) Selected() string {
	return m.selected
}
