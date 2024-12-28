package cloud

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCreateModel(t *testing.T) {
	model := NewCreateModel()

	assert.Equal(t, stepSelectProvider, model.currentStep)
	assert.NotNil(t, model.input)
	assert.NotNil(t, model.spinner)
}

func TestCreateModel_View(t *testing.T) {
	model := NewCreateModel()

	// Test view for stepSelectProvider
	model.currentStep = stepSelectProvider
	view := model.View()
	assert.Contains(t, view, "Select a cloud provider")

	// Test view for stepEnterConnectionName
	model.currentStep = stepEnterConnectionName
	view = model.View()
	assert.Contains(t, view, "Enter a name for the Cloud connection")

	// Test view for stepCreateSpinner
	model.currentStep = stepCreateSpinner
	view = model.View()
	assert.Contains(t, view, "Saving connection")

	// Test view for stepCreateDone
	model.currentStep = stepCreateDone
	view = model.View()
	assert.Contains(t, view, "Cloud connection created")
}

func TestCreateModel_ErrorHandling(t *testing.T) {
	model := NewCreateModel()

	// Simulate entering an empty connection name
	model.currentStep = stepEnterConnectionName
	model.input.SetValue("")
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	updatedModel, cmd := model.Update(msg)
	require.Nil(t, cmd)

	model = updatedModel.(*createModel)
	assert.Equal(t, stepEnterConnectionName, model.currentStep)
	assert.NotNil(t, model.err)
	assert.Contains(t, model.View(), "connection name can't be empty")
}

func TestCreateModel_Quit(t *testing.T) {
	model := NewCreateModel()

	// Simulate quitting at stepSelectProvider
	model.currentStep = stepSelectProvider
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	_, cmd := model.Update(msg)
	require.NotNil(t, cmd)
	assert.Equal(t, tea.QuitMsg{}, cmd())
}
