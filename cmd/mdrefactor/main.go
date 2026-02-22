package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/glebnaz/markdown-refactor/internal/diff"
	"github.com/glebnaz/markdown-refactor/internal/files"
	"github.com/glebnaz/markdown-refactor/internal/llm"
	"github.com/glebnaz/markdown-refactor/internal/tui"
)

type config struct {
	auto     bool
	endpoint string
	files    []string
}

func parseArgs(args []string) (config, error) {
	fs := flag.NewFlagSet("mdrefactor", flag.ContinueOnError)

	var cfg config
	fs.BoolVar(&cfg.auto, "auto", false, "Apply all changes without asking for approval")
	fs.StringVar(&cfg.endpoint, "endpoint", "http://localhost:1234", "LMStudio API endpoint")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	cfg.files = fs.Args()
	if len(cfg.files) == 0 {
		return config{}, fmt.Errorf("at least one file or directory argument is required")
	}

	return cfg, nil
}

type summary struct {
	fixed     int
	skipped   int
	unchanged int
	total     int
}

func (s summary) String() string {
	return fmt.Sprintf("%d file(s) fixed, %d skipped, %d unchanged", s.fixed, s.skipped, s.unchanged)
}

func processFile(ctx context.Context, w io.Writer, client llm.Client, filePath string, cfg config, index, total int) (result string, err error) {
	select {
	case <-ctx.Done():
		return "quit", ctx.Err()
	default:
	}

	shortName := filepath.Base(filePath)
	fmt.Fprintf(w, "Processing file %d/%d: %s\n", index, total, shortName)

	content, err := files.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", shortName, err)
	}

	corrected, err := client.Fix(ctx, content)
	if err != nil {
		return "", fmt.Errorf("fixing %s: %w", shortName, err)
	}

	diffResult := diff.Compute(shortName, content, corrected)

	if !diffResult.HasDiff {
		fmt.Fprintf(w, "  No changes needed for %s\n", shortName)
		return "unchanged", nil
	}

	if cfg.auto {
		fmt.Fprint(w, tui.RenderDiffPlain(diffResult))
		if err := files.WriteFile(filePath, corrected); err != nil {
			return "", fmt.Errorf("writing %s: %w", shortName, err)
		}
		fmt.Fprintf(w, "  Applied changes to %s\n", shortName)
		return "fixed", nil
	}

	// Interactive mode: show diff, then ask for approval.
	diffModel := tui.NewDiffModel(diffResult)
	p := tea.NewProgram(diffModel)
	if _, err := p.Run(); err != nil {
		return "", fmt.Errorf("diff viewer for %s: %w", shortName, err)
	}

	approveModel := tui.NewApproveModel(shortName)
	ap := tea.NewProgram(approveModel)
	finalApprove, err := ap.Run()
	if err != nil {
		return "", fmt.Errorf("approval prompt for %s: %w", shortName, err)
	}
	am := finalApprove.(tui.ApproveModel)

	switch am.Decision() {
	case tui.DecisionApprove:
		if err := files.WriteFile(filePath, corrected); err != nil {
			return "", fmt.Errorf("writing %s: %w", shortName, err)
		}
		fmt.Fprintf(w, "  Applied changes to %s\n", shortName)
		return "fixed", nil
	case tui.DecisionSkip:
		fmt.Fprintf(w, "  Skipped %s\n", shortName)
		return "skipped", nil
	case tui.DecisionQuit:
		return "quit", nil
	default:
		return "skipped", nil
	}
}

func execute(ctx context.Context, w io.Writer, cfg config, client llm.Client) error {
	resolvedFiles, err := files.Resolve(cfg.files)
	if err != nil {
		return fmt.Errorf("resolving files: %w", err)
	}

	s := summary{total: len(resolvedFiles)}

	for i, filePath := range resolvedFiles {
		result, err := processFile(ctx, w, client, filePath, cfg, i+1, s.total)
		if err != nil {
			return err
		}

		switch result {
		case "fixed":
			s.fixed++
		case "skipped":
			s.skipped++
		case "unchanged":
			s.unchanged++
		case "quit":
			fmt.Fprintf(w, "\nQuitting...\n")
			fmt.Fprintln(w, s)
			return nil
		}
	}

	fmt.Fprintf(w, "\n%s\n", s)
	return nil
}

func run() error {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := llm.NewClient(cfg.endpoint)
	return execute(ctx, os.Stdout, cfg, client)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
