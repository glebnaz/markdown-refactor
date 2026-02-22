package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/glebnaz/markdown-refactor/internal/diff"
)

func makeDiffResult(fileName string, lines []diff.LineChange) diff.Result {
	return diff.Result{
		FileName: fileName,
		Lines:    lines,
		HasDiff:  len(lines) > 0,
	}
}

func TestNewDiffModel(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "hello"},
	})
	m := NewDiffModel(result)

	if m.offset != 0 {
		t.Error("expected initial offset 0")
	}
	if m.quitting {
		t.Error("expected quitting to be false initially")
	}
}

func TestDiffModel_View_ShowsFileName(t *testing.T) {
	result := makeDiffResult("myfile.md", []diff.LineChange{
		{Type: diff.Equal, Text: "context line"},
	})
	m := NewDiffModel(result)
	view := m.View()

	if !strings.Contains(view, "myfile.md") {
		t.Error("expected file name in view output")
	}
}

func TestDiffModel_View_ShowsAddedLines(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Insert, Text: "new line"},
	})
	m := NewDiffModel(result)
	view := m.View()

	if !strings.Contains(view, "+ new line") {
		t.Errorf("expected '+ new line' in view, got: %s", view)
	}
}

func TestDiffModel_View_ShowsDeletedLines(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Delete, Text: "removed line"},
	})
	m := NewDiffModel(result)
	view := m.View()

	if !strings.Contains(view, "- removed line") {
		t.Errorf("expected '- removed line' in view, got: %s", view)
	}
}

func TestDiffModel_View_ShowsContextLines(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "context"},
	})
	m := NewDiffModel(result)
	view := m.View()

	if !strings.Contains(view, "  context") {
		t.Errorf("expected '  context' in view, got: %s", view)
	}
}

func TestDiffModel_View_ShowsHelp(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "line"},
	})
	m := NewDiffModel(result)
	view := m.View()

	if !strings.Contains(view, "scroll") {
		t.Error("expected help text with 'scroll' in view")
	}
}

func TestDiffModel_ScrollDown(t *testing.T) {
	lines := make([]diff.LineChange, 50)
	for i := range lines {
		lines[i] = diff.LineChange{Type: diff.Equal, Text: "line"}
	}
	result := makeDiffResult("test.md", lines)
	m := NewDiffModel(result)
	m.height = 10 // small viewport

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(DiffModel)

	if model.offset != 1 {
		t.Errorf("expected offset 1 after scroll down, got %d", model.offset)
	}
}

func TestDiffModel_ScrollUp(t *testing.T) {
	lines := make([]diff.LineChange, 50)
	for i := range lines {
		lines[i] = diff.LineChange{Type: diff.Equal, Text: "line"}
	}
	result := makeDiffResult("test.md", lines)
	m := NewDiffModel(result)
	m.height = 10
	m.offset = 5

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(DiffModel)

	if model.offset != 4 {
		t.Errorf("expected offset 4 after scroll up, got %d", model.offset)
	}
}

func TestDiffModel_ScrollUp_AtTop(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "line"},
	})
	m := NewDiffModel(result)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(DiffModel)

	if model.offset != 0 {
		t.Error("expected offset to stay at 0 when already at top")
	}
}

func TestDiffModel_ScrollDown_AtBottom(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "line"},
	})
	m := NewDiffModel(result)
	m.height = 10

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(DiffModel)

	if model.offset != 0 {
		t.Error("expected offset to stay at 0 when all content fits")
	}
}

func TestDiffModel_Quit(t *testing.T) {
	result := makeDiffResult("test.md", nil)
	m := NewDiffModel(result)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(DiffModel)

	if !model.Quitting() {
		t.Error("expected quitting to be true after 'q'")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestDiffModel_Enter(t *testing.T) {
	result := makeDiffResult("test.md", nil)
	m := NewDiffModel(result)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DiffModel)

	if !model.Quitting() {
		t.Error("expected quitting to be true after Enter")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestDiffModel_WindowResize(t *testing.T) {
	result := makeDiffResult("test.md", nil)
	m := NewDiffModel(result)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(DiffModel)

	if model.height != 40 {
		t.Errorf("expected height 40, got %d", model.height)
	}
	if model.width != 120 {
		t.Errorf("expected width 120, got %d", model.width)
	}
}

func TestDiffModel_VimKeys(t *testing.T) {
	lines := make([]diff.LineChange, 50)
	for i := range lines {
		lines[i] = diff.LineChange{Type: diff.Equal, Text: "line"}
	}
	result := makeDiffResult("test.md", lines)
	m := NewDiffModel(result)
	m.height = 10

	// j = down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model := updated.(DiffModel)
	if model.offset != 1 {
		t.Errorf("expected offset 1 after 'j', got %d", model.offset)
	}

	// k = up
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(DiffModel)
	if model.offset != 0 {
		t.Errorf("expected offset 0 after 'k', got %d", model.offset)
	}
}

func TestDiffModel_PageUpDown(t *testing.T) {
	lines := make([]diff.LineChange, 100)
	for i := range lines {
		lines[i] = diff.LineChange{Type: diff.Equal, Text: "line"}
	}
	result := makeDiffResult("test.md", lines)
	m := NewDiffModel(result)
	m.height = 12 // viewableLines = 10

	// Page down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	model := updated.(DiffModel)
	if model.offset != 10 {
		t.Errorf("expected offset 10 after PgDown, got %d", model.offset)
	}

	// Page up
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	model = updated.(DiffModel)
	if model.offset != 0 {
		t.Errorf("expected offset 0 after PgUp, got %d", model.offset)
	}
}

func TestDiffModel_HomeEnd(t *testing.T) {
	lines := make([]diff.LineChange, 100)
	for i := range lines {
		lines[i] = diff.LineChange{Type: diff.Equal, Text: "line"}
	}
	result := makeDiffResult("test.md", lines)
	m := NewDiffModel(result)
	m.height = 12

	// End (G)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	model := updated.(DiffModel)
	if model.offset != 90 { // 100 - 10 viewable
		t.Errorf("expected offset 90 after 'G', got %d", model.offset)
	}

	// Home (g)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	model = updated.(DiffModel)
	if model.offset != 0 {
		t.Errorf("expected offset 0 after 'g', got %d", model.offset)
	}
}

func TestDiffModel_Init_ReturnsNil(t *testing.T) {
	result := makeDiffResult("test.md", nil)
	m := NewDiffModel(result)
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil")
	}
}

func TestRenderDiffPlain(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Delete, Text: "old line"},
		{Type: diff.Insert, Text: "new line"},
		{Type: diff.Equal, Text: "context"},
	})

	output := RenderDiffPlain(result)

	if !strings.Contains(output, "test.md") {
		t.Error("expected file name in plain render")
	}
	if !strings.Contains(output, "- old line") {
		t.Error("expected delete marker in plain render")
	}
	if !strings.Contains(output, "+ new line") {
		t.Error("expected insert marker in plain render")
	}
	if !strings.Contains(output, "  context") {
		t.Error("expected context marker in plain render")
	}
}

func TestDiffModel_View_QuittingReturnsEmpty(t *testing.T) {
	result := makeDiffResult("test.md", []diff.LineChange{
		{Type: diff.Equal, Text: "line"},
	})
	m := NewDiffModel(result)
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Errorf("expected empty view when quitting, got: %s", view)
	}
}
