package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewApproveModel(t *testing.T) {
	m := NewApproveModel("test.md")

	if m.decision != DecisionPending {
		t.Error("expected initial decision to be DecisionPending")
	}
	if m.fileName != "test.md" {
		t.Errorf("expected fileName 'test.md', got %q", m.fileName)
	}
}

func TestApproveModel_Init_ReturnsNil(t *testing.T) {
	m := NewApproveModel("test.md")
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}

func TestApproveModel_View_ShowsPrompt(t *testing.T) {
	m := NewApproveModel("readme.md")
	view := m.View()

	if !strings.Contains(view, "readme.md") {
		t.Error("expected file name in view")
	}
	if !strings.Contains(view, "[Y]") {
		t.Error("expected [Y] option in view")
	}
	if !strings.Contains(view, "[N]") {
		t.Error("expected [N] option in view")
	}
	if !strings.Contains(view, "[Q]") {
		t.Error("expected [Q] option in view")
	}
}

func TestApproveModel_View_EmptyAfterDecision(t *testing.T) {
	m := NewApproveModel("test.md")
	m.decision = DecisionApprove

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view after decision, got: %s", view)
	}
}

func TestApproveModel_ApproveY(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionApprove {
		t.Errorf("expected DecisionApprove, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_ApproveUpperY(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionApprove {
		t.Errorf("expected DecisionApprove, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_SkipN(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionSkip {
		t.Errorf("expected DecisionSkip, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_SkipUpperN(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionSkip {
		t.Errorf("expected DecisionSkip, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_QuitQ(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionQuit {
		t.Errorf("expected DecisionQuit, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_QuitUpperQ(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionQuit {
		t.Errorf("expected DecisionQuit, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_QuitCtrlC(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionQuit {
		t.Errorf("expected DecisionQuit, got %v", model.Decision())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestApproveModel_IgnoresUnknownKeys(t *testing.T) {
	m := NewApproveModel("test.md")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(ApproveModel)

	if model.Decision() != DecisionPending {
		t.Errorf("expected DecisionPending after unknown key, got %v", model.Decision())
	}
	if cmd != nil {
		t.Error("expected no command for unknown key")
	}
}

func TestApproveModel_DecisionMethod(t *testing.T) {
	m := NewApproveModel("test.md")

	if m.Decision() != DecisionPending {
		t.Errorf("expected DecisionPending, got %v", m.Decision())
	}

	m.decision = DecisionApprove
	if m.Decision() != DecisionApprove {
		t.Errorf("expected DecisionApprove, got %v", m.Decision())
	}

	m.decision = DecisionSkip
	if m.Decision() != DecisionSkip {
		t.Errorf("expected DecisionSkip, got %v", m.Decision())
	}

	m.decision = DecisionQuit
	if m.Decision() != DecisionQuit {
		t.Errorf("expected DecisionQuit, got %v", m.Decision())
	}
}
