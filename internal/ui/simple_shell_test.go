package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/prompt-ops/pops/internal/domain"
	"github.com/prompt-ops/pops/internal/services"
)

// MockShellService for testing
type MockShellService struct{}

func (m *MockShellService) StartSession(ctx context.Context, connectionName string) (*domain.Session, error) {
	return &domain.Session{ID: "test-session"}, nil
}

func (m *MockShellService) ProcessInput(ctx context.Context, connectionName string, input string, mode domain.QueryMode) (*services.ShellResponse, error) {
	if mode == domain.QueryModeCommand {
		return &services.ShellResponse{
			Type:    services.ResponseTypeCommand,
			Command: "SELECT * FROM users WHERE id = 1;",
		}, nil
	}
	return &services.ShellResponse{
		Type:   services.ResponseTypeAnswer,
		Answer: "This is a test answer.",
	}, nil
}

func (m *MockShellService) ExecuteCommand(ctx context.Context, sessionID domain.SessionID, commandID domain.CommandID) (*domain.ExecutionResult, error) {
	return &domain.ExecutionResult{
		CommandID: commandID,
		Status:    domain.ExecutionStatusCompleted,
		Output:    "Mock execution result",
	}, nil
}

func (m *MockShellService) ShowContext(ctx context.Context, connectionName string) (*services.ContextDisplay, error) {
	return &services.ContextDisplay{
		Title: "Test Context",
		Items: []services.ContextItem{
			{Name: "users", Type: "table"},
			{Name: "orders", Type: "table"},
		},
	}, nil
}

func TestSeparateHeaderAndData(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})

	tests := []struct {
		name           string
		input          []string
		expectedHeader []string
		expectedData   []string
	}{
		{
			name: "Standard ASCII table with clear headers",
			input: []string{
				"+----+------------------+------------------+",
				"| ID | FIRST_NAME       | LAST_NAME        |",
				"+----+------------------+------------------+",
				"| 1  | John             | Smith            |",
				"| 2  | Sarah            | Johnson          |",
				"| 3  | Michael          | Brown            |",
			},
			expectedHeader: []string{
				"+----+------------------+------------------+",
				"| ID | FIRST_NAME       | LAST_NAME        |",
				"+----+------------------+------------------+",
			},
			expectedData: []string{
				"| 1  | John             | Smith            |",
				"| 2  | Sarah            | Johnson          |",
				"| 3  | Michael          | Brown            |",
			},
		},
		{
			name: "Table with EMAIL and DEPARTMENT headers",
			input: []string{
				"+----+----------+--------------------+",
				"| ID | NAME     | EMAIL              |",
				"+----+----------+--------------------+",
				"| 1  | Alice    | alice@company.com  |",
				"| 2  | Bob      | bob@company.com    |",
			},
			expectedHeader: []string{
				"+----+----------+--------------------+",
				"| ID | NAME     | EMAIL              |",
				"+----+----------+--------------------+",
			},
			expectedData: []string{
				"| 1  | Alice    | alice@company.com  |",
				"| 2  | Bob      | bob@company.com    |",
			},
		},
		{
			name: "No clear table structure - should treat as all data",
			input: []string{
				"Some random text",
				"More random text",
				"Not a table at all",
			},
			expectedHeader: []string{},
			expectedData: []string{
				"Some random text",
				"More random text",
				"Not a table at all",
			},
		},
		{
			name: "Table with numeric data that looks like headers",
			input: []string{
				"+----+----------+----------+",
				"| ID | NAME     | SALARY   |",
				"+----+----------+----------+",
				"| 1  | Employee | 50000    |",
				"| 2  | Manager  | 75000    |",
				"| 3  | Director | 100000   |",
			},
			expectedHeader: []string{
				"+----+----------+----------+",
				"| ID | NAME     | SALARY   |",
				"+----+----------+----------+",
			},
			expectedData: []string{
				"| 1  | Employee | 50000    |",
				"| 2  | Manager  | 75000    |",
				"| 3  | Director | 100000   |",
			},
		},
		{
			name: "Empty input",
			input: []string{},
			expectedHeader: []string{},
			expectedData: []string{},
		},
		{
			name: "Single line",
			input: []string{"| ID | NAME |"},
			expectedHeader: []string{},
			expectedData: []string{"| ID | NAME |"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headerLines, dataLines := shell.separateHeaderAndData(tt.input)

			// Check headers
			if len(headerLines) != len(tt.expectedHeader) {
				t.Errorf("Header length mismatch. Expected %d, got %d", len(tt.expectedHeader), len(headerLines))
			}
			for i, expected := range tt.expectedHeader {
				if i < len(headerLines) && headerLines[i] != expected {
					t.Errorf("Header line %d mismatch. Expected %q, got %q", i, expected, headerLines[i])
				}
			}

			// Check data
			if len(dataLines) != len(tt.expectedData) {
				t.Errorf("Data length mismatch. Expected %d, got %d", len(tt.expectedData), len(dataLines))
			}
			for i, expected := range tt.expectedData {
				if i < len(dataLines) && dataLines[i] != expected {
					t.Errorf("Data line %d mismatch. Expected %q, got %q", i, expected, dataLines[i])
				}
			}
		})
	}
}

func TestRenderScrollableOutput(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})
	
	// Set up test dimensions
	shell.width = 50
	shell.height = 20

	tests := []struct {
		name           string
		input          string
		setupScrolling func(*SimpleShell)
		expectedLines  int
		description    string
	}{
		{
			name: "Normal table that fits",
			input: `+----+------+
| ID | NAME |
+----+------+
| 1  | John |
| 2  | Jane |
+----+------+`,
			setupScrolling: func(s *SimpleShell) {
				// No scrolling needed
			},
			expectedLines: 6,
			description:   "Should show all lines when content fits",
		},
		{
			name: "Wide table requiring horizontal scroll",
			input: `+----+------------------+------------------+------------------+
| ID | VERY_LONG_COLUMN | ANOTHER_LONG_COL | THIRD_LONG_COL   |
+----+------------------+------------------+------------------+
| 1  | Very long data   | More long data   | Even more data   |
+----+------------------+------------------+------------------+`,
			setupScrolling: func(s *SimpleShell) {
				s.scrollOffset = 10 // Scroll right by 10 characters
			},
			expectedLines: 5,
			description:   "Should apply horizontal scrolling",
		},
		{
			name: "Tall table requiring vertical scroll",
			input: strings.Join([]string{
				"+----+------+",
				"| ID | NAME |",
				"+----+------+",
				"| 1  | John |",
				"| 2  | Jane |",
				"| 3  | Bob  |",
				"| 4  | Alice|",
				"| 5  | Tom  |",
				"| 6  | Lisa |",
				"| 7  | Mike |",
				"| 8  | Sara |",
				"| 9  | Dave |",
				"| 10 | Emma |",
				"| 11 | Jack |",
				"| 12 | Jill |",
				"| 13 | Paul |",
				"| 14 | Rose |",
				"| 15 | Mark |",
				"+----+------+",
			}, "\n"),
			setupScrolling: func(s *SimpleShell) {
				s.verticalOffset = 5 // Scroll down by 5 lines
			},
			expectedLines: 8, // Should be limited by available height
			description:   "Should apply vertical scrolling with sticky headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup scrolling state
			tt.setupScrolling(shell)

			// Render scrollable output
			result := shell.renderScrollableOutput(tt.input)
			resultLines := strings.Split(result, "\n")

			// Check line count
			if len(resultLines) != tt.expectedLines {
				t.Errorf("%s: Expected %d lines, got %d", tt.description, tt.expectedLines, len(resultLines))
			}

			// Verify it's a non-empty result for valid inputs
			if tt.input != "" && result == "" {
				t.Errorf("%s: Expected non-empty result for non-empty input", tt.description)
			}
		})
	}
}

func TestScrollBounds(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})
	shell.width = 30
	shell.height = 10

	// Test with content that requires both horizontal and vertical scrolling
	wideAndTallContent := strings.Join([]string{
		"+----+------------------+------------------+------------------+",
		"| ID | VERY_LONG_COLUMN | ANOTHER_LONG_COL | THIRD_LONG_COL   |",
		"+----+------------------+------------------+------------------+",
		"| 1  | Very long data 1 | More long data 1 | Even more data 1 |",
		"| 2  | Very long data 2 | More long data 2 | Even more data 2 |",
		"| 3  | Very long data 3 | More long data 3 | Even more data 3 |",
		"| 4  | Very long data 4 | More long data 4 | Even more data 4 |",
		"| 5  | Very long data 5 | More long data 5 | Even more data 5 |",
		"| 6  | Very long data 6 | More long data 6 | Even more data 6 |",
		"| 7  | Very long data 7 | More long data 7 | Even more data 7 |",
		"| 8  | Very long data 8 | More long data 8 | Even more data 8 |",
		"+----+------------------+------------------+------------------+",
	}, "\n")

	// Render to calculate scroll bounds
	shell.renderScrollableOutput(wideAndTallContent)

	tests := []struct {
		name        string
		testValue   int
		maxValue    int
		description string
	}{
		{
			name:        "Horizontal scroll bounds",
			testValue:   shell.maxScroll,
			maxValue:    0, // Should be > 0 since content is wider than available width
			description: "Should calculate horizontal scroll bounds for wide content",
		},
		{
			name:        "Vertical scroll bounds", 
			testValue:   shell.maxVertical,
			maxValue:    0, // Should be > 0 since content is taller than available height
			description: "Should calculate vertical scroll bounds for tall content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testValue <= tt.maxValue {
				t.Errorf("%s: Expected %s to be > %d, got %d", tt.description, tt.name, tt.maxValue, tt.testValue)
			}
		})
	}
}

func TestResetToInput(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})
	
	// Set some non-default values
	shell.scrollOffset = 10
	shell.maxScroll = 20
	shell.verticalOffset = 5
	shell.maxVertical = 15
	shell.confirmScrollOffset = 3
	shell.confirmMaxScroll = 8
	shell.confirmVertOffset = 2
	shell.confirmMaxVertical = 6
	shell.historyIndex = 1
	
	// Reset to input
	shell.resetToInput()
	
	// Verify all scroll values are reset
	if shell.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to be 0, got %d", shell.scrollOffset)
	}
	if shell.maxScroll != 0 {
		t.Errorf("Expected maxScroll to be 0, got %d", shell.maxScroll)
	}
	if shell.verticalOffset != 0 {
		t.Errorf("Expected verticalOffset to be 0, got %d", shell.verticalOffset)
	}
	if shell.maxVertical != 0 {
		t.Errorf("Expected maxVertical to be 0, got %d", shell.maxVertical)
	}
	if shell.confirmScrollOffset != 0 {
		t.Errorf("Expected confirmScrollOffset to be 0, got %d", shell.confirmScrollOffset)
	}
	if shell.confirmMaxScroll != 0 {
		t.Errorf("Expected confirmMaxScroll to be 0, got %d", shell.confirmMaxScroll)
	}
	if shell.confirmVertOffset != 0 {
		t.Errorf("Expected confirmVertOffset to be 0, got %d", shell.confirmVertOffset)
	}
	if shell.confirmMaxVertical != 0 {
		t.Errorf("Expected confirmMaxVertical to be 0, got %d", shell.confirmMaxVertical)
	}
	if shell.historyIndex != 0 {
		t.Errorf("Expected historyIndex to be 0, got %d", shell.historyIndex)
	}
}

func TestAddToHistory(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})
	
	// Test adding history entries
	shell.addToHistory("test prompt 1", "test response 1", true)
	shell.addToHistory("test prompt 2", "test response 2", false)
	
	if len(shell.history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(shell.history))
	}
	
	// Check first entry
	if shell.history[0].Prompt != "test prompt 1" {
		t.Errorf("Expected first prompt to be 'test prompt 1', got %q", shell.history[0].Prompt)
	}
	if shell.history[0].Success != true {
		t.Errorf("Expected first entry to be successful")
	}
	
	// Check second entry
	if shell.history[1].Prompt != "test prompt 2" {
		t.Errorf("Expected second prompt to be 'test prompt 2', got %q", shell.history[1].Prompt)
	}
	if shell.history[1].Success != false {
		t.Errorf("Expected second entry to be unsuccessful")
	}
	
	// Test history limit (should keep only last 10 items)
	for i := 3; i <= 15; i++ {
		shell.addToHistory(fmt.Sprintf("prompt %d", i), fmt.Sprintf("response %d", i), true)
	}
	
	if len(shell.history) != 10 {
		t.Errorf("Expected history to be limited to 10 entries, got %d", len(shell.history))
	}
	
	// Should have entries 6-15 (last 10)
	if !strings.Contains(shell.history[0].Prompt, "prompt 6") {
		t.Errorf("Expected oldest entry to be 'prompt 6', got %q", shell.history[0].Prompt)
	}
	if !strings.Contains(shell.history[9].Prompt, "prompt 15") {
		t.Errorf("Expected newest entry to be 'prompt 15', got %q", shell.history[9].Prompt)
	}
}

func TestMoveHistoryItemToTop(t *testing.T) {
	shell := NewSimpleShell("test-conn", &MockShellService{})
	
	// Add some history entries
	shell.addToHistory("prompt A", "response A", true)
	shell.addToHistory("prompt B", "response B", true)
	shell.addToHistory("prompt C", "response C", true)
	
	// History should be [A, B, C] (C is most recent)
	if shell.history[0].Prompt != "prompt A" || shell.history[2].Prompt != "prompt C" {
		t.Errorf("Initial history order incorrect")
	}
	
	// Move item 0 (A) to top
	shell.moveHistoryItemToTop(0)
	
	// History should now be [B, C, A] (A is most recent)
	if len(shell.history) != 3 {
		t.Errorf("Expected 3 history entries after move, got %d", len(shell.history))
	}
	
	if shell.history[0].Prompt != "prompt B" {
		t.Errorf("Expected first entry to be 'prompt B', got %q", shell.history[0].Prompt)
	}
	if shell.history[1].Prompt != "prompt C" {
		t.Errorf("Expected second entry to be 'prompt C', got %q", shell.history[1].Prompt)
	}
	if shell.history[2].Prompt != "prompt A" {
		t.Errorf("Expected third entry (most recent) to be 'prompt A', got %q", shell.history[2].Prompt)
	}
	
	// Test moving middle item (C) to top
	shell.moveHistoryItemToTop(1)
	
	// History should now be [B, A, C] (C is most recent again)
	if shell.history[0].Prompt != "prompt B" {
		t.Errorf("Expected first entry to be 'prompt B', got %q", shell.history[0].Prompt)
	}
	if shell.history[1].Prompt != "prompt A" {
		t.Errorf("Expected second entry to be 'prompt A', got %q", shell.history[1].Prompt)
	}
	if shell.history[2].Prompt != "prompt C" {
		t.Errorf("Expected third entry (most recent) to be 'prompt C', got %q", shell.history[2].Prompt)
	}
	
	// Test invalid indices
	shell.moveHistoryItemToTop(-1) // Should not crash
	shell.moveHistoryItemToTop(10) // Should not crash
	
	// History should remain unchanged
	if len(shell.history) != 3 {
		t.Errorf("Expected history to remain unchanged after invalid moves")
	}
}