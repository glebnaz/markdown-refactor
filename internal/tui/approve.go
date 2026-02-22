package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Decision represents the user's choice in the approval prompt.
type Decision int

const (
	// DecisionPending means the user hasn't decided yet.
	DecisionPending Decision = iota
	// DecisionApprove means apply the changes.
	DecisionApprove
	// DecisionSkip means skip this file.
	DecisionSkip
	// DecisionQuit means quit the entire program.
	DecisionQuit
)

var (
	promptStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	highlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
)

// ApproveModel is a Bubble Tea model for the approve/skip/quit prompt.
type ApproveModel struct {
	fileName string
	decision Decision
}

// NewApproveModel creates a new approval prompt for the given file.
func NewApproveModel(fileName string) ApproveModel {
	return ApproveModel{
		fileName: fileName,
		decision: DecisionPending,
	}
}

// Init implements tea.Model.
func (m ApproveModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ApproveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.decision = DecisionApprove
			return m, tea.Quit
		case "n", "N":
			m.decision = DecisionSkip
			return m, tea.Quit
		case "q", "Q", "ctrl+c":
			m.decision = DecisionQuit
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m ApproveModel) View() string {
	if m.decision != DecisionPending {
		return ""
	}
	prompt := promptStyle.Render(fmt.Sprintf("Apply changes to %s?", m.fileName))
	options := fmt.Sprintf("  %s apply  %s skip  %s quit",
		highlightStyle.Render("[Y]"),
		highlightStyle.Render("[N]"),
		highlightStyle.Render("[Q]"),
	)
	return prompt + "\n" + options + "\n"
}

// Decision returns the user's decision.
func (m ApproveModel) Decision() Decision {
	return m.decision
}
