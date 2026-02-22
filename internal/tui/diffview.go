package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/glebnaz/markdown-refactor/internal/diff"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	addStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	deleteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	contextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
)

// DiffModel is a Bubble Tea model that displays a scrollable colored diff.
type DiffModel struct {
	result   diff.Result
	offset   int
	height   int
	width    int
	quitting bool
}

// NewDiffModel creates a new DiffModel for the given diff result.
func NewDiffModel(result diff.Result) DiffModel {
	return DiffModel{
		result: result,
		offset: 0,
		height: 24, // default, updated by WindowSizeMsg
		width:  80,
	}
}

// Init implements tea.Model.
func (m DiffModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			maxOffset := m.maxOffset()
			if m.offset < maxOffset {
				m.offset++
			}
		case "pgup":
			m.offset -= m.viewableLines()
			if m.offset < 0 {
				m.offset = 0
			}
		case "pgdown":
			m.offset += m.viewableLines()
			maxOffset := m.maxOffset()
			if m.offset > maxOffset {
				m.offset = maxOffset
			}
		case "home", "g":
			m.offset = 0
		case "end", "G":
			m.offset = m.maxOffset()
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m DiffModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	header := headerStyle.Render(fmt.Sprintf("--- %s", m.result.FileName))
	b.WriteString(header)
	b.WriteString("\n")

	viewable := m.viewableLines()
	lines := m.result.Lines

	end := m.offset + viewable
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[m.offset:end]
	for _, line := range visible {
		rendered := renderLine(line)
		b.WriteString(rendered)
		b.WriteString("\n")
	}

	// Pad remaining space.
	rendered := end - m.offset
	for i := rendered; i < viewable; i++ {
		b.WriteString("\n")
	}

	help := helpStyle.Render("↑/↓/j/k: scroll  PgUp/PgDn: page  Enter: continue  q: quit")
	b.WriteString(help)

	return b.String()
}

// Quitting returns true if the user chose to quit.
func (m DiffModel) Quitting() bool {
	return m.quitting
}

// viewableLines returns how many diff lines fit in the viewport.
// We subtract 2 for the header and help lines.
func (m DiffModel) viewableLines() int {
	v := m.height - 2
	if v < 1 {
		v = 1
	}
	return v
}

// maxOffset returns the maximum scroll offset.
func (m DiffModel) maxOffset() int {
	max := len(m.result.Lines) - m.viewableLines()
	if max < 0 {
		return 0
	}
	return max
}

func renderLine(line diff.LineChange) string {
	switch line.Type {
	case diff.Insert:
		return addStyle.Render("+ " + line.Text)
	case diff.Delete:
		return deleteStyle.Render("- " + line.Text)
	default:
		return contextStyle.Render("  " + line.Text)
	}
}

// RenderDiffPlain renders a diff result as plain colored text (non-interactive).
// Useful for --auto mode or piped output.
func RenderDiffPlain(result diff.Result) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("--- %s", result.FileName)))
	b.WriteString("\n")
	for _, line := range result.Lines {
		b.WriteString(renderLine(line))
		b.WriteString("\n")
	}
	return b.String()
}
