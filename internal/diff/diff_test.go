package diff

import (
	"strings"
	"testing"
)

func TestCompute_NoChanges(t *testing.T) {
	original := "Hello, world!\nThis is a test.\n"
	result := Compute("test.md", original, original)

	if result.HasDiff {
		t.Error("expected HasDiff to be false when texts are identical")
	}
	if result.FileName != "test.md" {
		t.Errorf("expected FileName 'test.md', got '%s'", result.FileName)
	}
	if len(result.Lines) != 0 {
		t.Errorf("expected no lines, got %d", len(result.Lines))
	}
}

func TestCompute_Additions(t *testing.T) {
	original := "Line one.\n"
	corrected := "Line one.\nLine two.\n"
	result := Compute("test.md", original, corrected)

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
	if !hasChangeType(result.Lines, Insert) {
		t.Error("expected at least one Insert line")
	}
}

func TestCompute_Removals(t *testing.T) {
	original := "Line one.\nLine two.\n"
	corrected := "Line one.\n"
	result := Compute("test.md", original, corrected)

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
	if !hasChangeType(result.Lines, Delete) {
		t.Error("expected at least one Delete line")
	}
}

func TestCompute_Mixed(t *testing.T) {
	original := "Ths is a tset.\nKeep this line.\nRemove me.\n"
	corrected := "This is a test.\nKeep this line.\nAdded line.\n"
	result := Compute("doc.md", original, corrected)

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
	if !hasChangeType(result.Lines, Insert) {
		t.Error("expected Insert lines in mixed diff")
	}
	if !hasChangeType(result.Lines, Delete) {
		t.Error("expected Delete lines in mixed diff")
	}
}

func TestCompute_EmptyOriginal(t *testing.T) {
	result := Compute("empty.md", "", "Some content.\n")

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
	if !hasChangeType(result.Lines, Insert) {
		t.Error("expected Insert lines when original is empty")
	}
}

func TestCompute_EmptyCorrected(t *testing.T) {
	result := Compute("empty.md", "Some content.\n", "")

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
	if !hasChangeType(result.Lines, Delete) {
		t.Error("expected Delete lines when corrected is empty")
	}
}

func TestCompute_BothEmpty(t *testing.T) {
	result := Compute("empty.md", "", "")

	if result.HasDiff {
		t.Error("expected HasDiff to be false when both are empty")
	}
}

func TestCompute_GrammarFix(t *testing.T) {
	original := "He dont like apples.\n"
	corrected := "He doesn't like apples.\n"
	result := Compute("grammar.md", original, corrected)

	if !result.HasDiff {
		t.Error("expected HasDiff to be true for grammar fix")
	}
}

func TestCompute_MultilineChanges(t *testing.T) {
	original := "# Title\n\nFirst paragrph with error.\n\nSecond paragraph is fine.\n\nThrid paragraph also has error.\n"
	corrected := "# Title\n\nFirst paragraph with error.\n\nSecond paragraph is fine.\n\nThird paragraph also has error.\n"
	result := Compute("multi.md", original, corrected)

	if !result.HasDiff {
		t.Error("expected HasDiff to be true")
	}
}

func TestFormatUnified_NoChanges(t *testing.T) {
	result := Result{
		FileName: "test.md",
		HasDiff:  false,
	}
	output := FormatUnified(result)
	if !strings.Contains(output, "No changes") {
		t.Error("expected 'No changes' in output for no-diff result")
	}
	if !strings.Contains(output, "test.md") {
		t.Error("expected filename in output")
	}
}

func TestFormatUnified_WithChanges(t *testing.T) {
	result := Compute("test.md", "old line\n", "new line\n")
	output := FormatUnified(result)

	if !strings.Contains(output, "--- test.md") {
		t.Error("expected --- header in unified output")
	}
	if !strings.Contains(output, "+++ test.md") {
		t.Error("expected +++ header in unified output")
	}
	if !strings.Contains(output, "- ") {
		t.Error("expected deletion line with '- ' prefix")
	}
	if !strings.Contains(output, "+ ") {
		t.Error("expected addition line with '+ ' prefix")
	}
}

func TestFormatUnified_EqualLines(t *testing.T) {
	result := Compute("test.md", "same\ndifferent\n", "same\nchanged\n")
	output := FormatUnified(result)

	if !strings.Contains(output, "  same") {
		t.Error("expected equal/context lines with '  ' prefix")
	}
}

func hasChangeType(lines []LineChange, ct ChangeType) bool {
	for _, l := range lines {
		if l.Type == ct {
			return true
		}
	}
	return false
}
