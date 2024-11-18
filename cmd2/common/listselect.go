package common

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ListSelectModel struct {
	list     list.Model
	choices  []string
	selected string
	err      error
}

type listItem struct {
	title string
}

func (i listItem) Title() string {
	return i.title
}

func (i listItem) Description() string {
	return ""
}

func (i listItem) FilterValue() string {
	return i.title
}

func NewListSelectModel(title string, choices []string) ListSelectModel {
	items := make([]list.Item, len(choices))
	for i, choice := range choices {
		items[i] = listItem{
			title: choice,
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title

	return ListSelectModel{
		list:    l,
		choices: choices,
	}
}

func (m ListSelectModel) Init() tea.Cmd {
	return nil
}

func (m ListSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.selected = m.choices[m.list.Index()]
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListSelectModel) View() string {
	return m.list.View()
}

func (m ListSelectModel) Selected() string {
	return m.selected
}
