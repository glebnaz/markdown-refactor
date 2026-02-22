package diff

import (
	"fmt"
	"strings"

	difflib "github.com/sergi/go-diff/diffmatchpatch"
)

// LineChange represents a single line in a diff with its type.
type LineChange struct {
	Type ChangeType
	Text string
}

// ChangeType indicates whether a line was added, removed, or unchanged.
type ChangeType int

const (
	Equal  ChangeType = iota
	Insert
	Delete
)

// Result holds the diff between two texts.
type Result struct {
	FileName string
	Lines    []LineChange
	HasDiff  bool
}

// Compute computes a line-by-line diff between original and corrected text.
// It returns a Result with the diff lines and whether any changes were found.
func Compute(fileName, original, corrected string) Result {
	if original == corrected {
		return Result{
			FileName: fileName,
			Lines:    nil,
			HasDiff:  false,
		}
	}

	dmp := difflib.New()
	diffs := dmp.DiffMain(original, corrected, true)
	diffs = dmp.DiffCleanupSemantic(diffs)

	lines := diffToLines(diffs)

	return Result{
		FileName: fileName,
		Lines:    lines,
		HasDiff:  true,
	}
}

// diffToLines converts character-level diffs into line-level changes.
func diffToLines(diffs []difflib.Diff) []LineChange {
	var lines []LineChange

	for _, d := range diffs {
		ct := toChangeType(d.Type)
		parts := strings.Split(d.Text, "\n")
		for i, part := range parts {
			if part == "" && i == len(parts)-1 {
				continue
			}
			lines = append(lines, LineChange{
				Type: ct,
				Text: part,
			})
		}
	}

	return lines
}

func toChangeType(dt difflib.Operation) ChangeType {
	switch dt {
	case difflib.DiffInsert:
		return Insert
	case difflib.DiffDelete:
		return Delete
	default:
		return Equal
	}
}

// FormatUnified returns a unified-diff-style string representation of the result.
func FormatUnified(r Result) string {
	if !r.HasDiff {
		return fmt.Sprintf("--- %s\n+++ %s\nNo changes.\n", r.FileName, r.FileName)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", r.FileName, r.FileName))
	for _, line := range r.Lines {
		switch line.Type {
		case Delete:
			b.WriteString("- " + line.Text + "\n")
		case Insert:
			b.WriteString("+ " + line.Text + "\n")
		case Equal:
			b.WriteString("  " + line.Text + "\n")
		}
	}
	return b.String()
}
