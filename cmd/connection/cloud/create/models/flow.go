package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/prompt-ops/cli/pkg/azure"
	"github.com/prompt-ops/cli/pkg/common"
	config "github.com/prompt-ops/cli/pkg/config"
)

var providerMap = map[string]func(string) common.CloudConnection{
	"azure": func(name string) common.CloudConnection { return azure.NewAzureConnection(name) },
}

const (
	stepSelectProvider = iota
	stepEnterName
	stepEnd
)

type flowModel struct {
	step                   int
	provider               string
	name                   string
	err                    error
	quitting               bool
	providerSelectionModel providerSelectionModel
	enterNameModel         enterNameModel
	endModel               endModel
}

func initialFlowModel() flowModel {
	return flowModel{
		step:                   stepSelectProvider,
		providerSelectionModel: NewProviderSelectionModel(),
		enterNameModel:         NewEnterNameModel(),
		endModel:               NewEndModel(),
	}
}

func (m flowModel) Init() tea.Cmd {
	return nil
}

func (m flowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.step {
	case stepSelectProvider:
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.providerSelectionModel.Update(msg)
		m.providerSelectionModel = model.(providerSelectionModel)
		if m.providerSelectionModel.Quitting() {
			m.quitting = true
			return m, tea.Quit
		}
		if m.providerSelectionModel.Choice() != "" {
			m.provider = m.providerSelectionModel.Choice()
			m.step = stepEnterName
			return m, m.enterNameModel.Init()
		}
		return m, cmd

	case stepEnterName:
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.enterNameModel.Update(msg)
		m.enterNameModel = model.(enterNameModel)
		if m.enterNameModel.Quitting() {
			m.quitting = true
			return m, tea.Quit
		}
		if m.enterNameModel.Done() {
			m.name = m.enterNameModel.Value()

			nameExists := config.CheckIfNameExists(m.name)
			if !nameExists {
				m.err = fmt.Errorf("A connection with the name %q already exists", m.name)
				m.enterNameModel = NewEnterNameModel()
				return m, nil
			}

			m.step = stepEnd
			return m, m.endModel.Init()
		}
		return m, cmd

	case stepEnd:
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.endModel.Update(msg)
		m.endModel = model.(endModel)
		if m.endModel.Quitting() {
			m.quitting = true
			return m, tea.Quit
		}
		if m.endModel.Done() {
			return m, tea.Quit
		}
		return m, cmd

	default:
		return m, tea.Quit
	}
}

func (m flowModel) View() string {
	if m.quitting {
		return "\nExiting...\n"
	}
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}
	switch m.step {
	case stepSelectProvider:
		return m.providerSelectionModel.View()
	case stepEnterName:
		return m.enterNameModel.View()
	case stepEnd:
		return m.endModel.View()
	default:
		return ""
	}
}

func NewFlowModel() flowModel {
	return initialFlowModel()
}
