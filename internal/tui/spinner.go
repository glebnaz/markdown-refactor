package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerResult holds the outcome of the background work.
type SpinnerResult struct {
	Value string
	Err   error
}

type spinnerDoneMsg SpinnerResult

// SpinnerModel shows a spinner while a background function runs.
type SpinnerModel struct {
	spinner  spinner.Model
	message  string
	done     bool
	result   SpinnerResult
	workFunc func() (string, error)
}

// NewSpinnerModel creates a spinner that displays message while workFunc runs.
func NewSpinnerModel(message string, workFunc func() (string, error)) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return SpinnerModel{
		spinner:  s,
		message:  message,
		workFunc: workFunc,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.workFunc()
			return spinnerDoneMsg{Value: result, Err: err}
		},
	)
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.result = SpinnerResult(msg)
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			m.result = SpinnerResult{Err: fmt.Errorf("interrupted")}
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + m.message + "\n"
}

// Result returns the outcome of the background work.
func (m SpinnerModel) Result() SpinnerResult {
	return m.result
}
