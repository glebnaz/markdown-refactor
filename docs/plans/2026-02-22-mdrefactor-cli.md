# mdrefactor - CLI tool for Markdown grammar correction

## Overview
- CLI utility `mdrefactor` that reads Markdown files and fixes grammar, spelling, and syntax errors without changing meaning
- Uses local LMStudio (OpenAI-compatible API) as the LLM backend
- Shows colored diff of changes and asks for user approval before applying
- Supports `--auto` flag to skip approval prompts
- Supports multiple files and directory input
- Built with Bubble Tea (charmbracelet) for beautiful terminal UI
- Written in Go

## Context (from discovery)
- Fresh repository, no existing code
- Remote: `git@github.com:glebnaz/markdown-refactor.git`
- Language: Go
- TUI: charmbracelet/bubbletea + lipgloss + glamour
- LLM: LMStudio local server (OpenAI-compatible API at localhost:1234)

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change

## Testing Strategy
- **Unit tests**: required for every task
- Mock LMStudio API responses for testing LLM client
- Use testable architecture (interfaces for LLM client, file I/O)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with warning prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Initialize Go project and set up project structure
- [x] run `go mod init github.com/glebnaz/markdown-refactor`
- [x] create directory structure: `cmd/mdrefactor/`, `internal/llm/`, `internal/diff/`, `internal/tui/`
- [x] create `cmd/mdrefactor/main.go` with basic CLI flag parsing (file/dir args, `--auto` flag, `--endpoint` flag for LMStudio URL)
- [x] use standard `flag` package for CLI args (keep it simple)
- [x] write tests for flag parsing and argument validation
- [x] run tests - must pass before next task

### Task 2: Implement LMStudio client
- [x] create `internal/llm/client.go` with `Client` interface and `LMStudioClient` struct
- [x] implement `Fix(ctx context.Context, content string) (string, error)` method
- [x] use OpenAI-compatible chat completions API (`POST /v1/chat/completions`)
- [x] craft system prompt: "Fix grammar, spelling, and syntax errors in the following Markdown. Do not change the meaning. Return only the corrected Markdown, no explanations."
- [x] default endpoint: `http://localhost:1234`
- [x] configurable via `--endpoint` flag
- [x] write tests with HTTP test server mocking LMStudio responses
- [x] write tests for error cases (connection refused, bad response, empty response)
- [x] run tests - must pass before next task

### Task 3: Implement file reading and multi-file support
- [x] create `internal/files/reader.go` with functions to resolve input paths
- [x] support: single file, multiple files, directory (recursively find `*.md` files)
- [x] validate files exist and are readable
- [x] write tests for single file, multiple files, directory resolution
- [x] write tests for error cases (file not found, not a markdown file, empty directory)
- [x] run tests - must pass before next task

### Task 4: Implement diff generation and display
- [x] create `internal/diff/diff.go` with function to compute line-by-line diff between original and corrected text
- [x] use a Go diff library (e.g., `github.com/sergi/go-diff`) for unified diff
- [x] create `internal/tui/diffview.go` - Bubble Tea model for scrollable colored diff display
- [x] use lipgloss for styling: green for additions, red for removals, gray for context
- [x] show file name header above each diff
- [x] write tests for diff computation (no changes, additions, removals, mixed)
- [x] run tests - must pass before next task

### Task 5: Implement approval flow and file writing
- [ ] create `internal/tui/approve.go` - Bubble Tea model for approve/skip/quit prompt after diff view
- [ ] options: [Y]es apply, [N]o skip, [Q]uit
- [ ] when approved, write corrected content back to original file
- [ ] when `--auto` flag is set, skip approval and apply all changes automatically
- [ ] if no changes detected for a file, skip it with a message
- [ ] write tests for file writing logic (apply, skip scenarios)
- [ ] write tests for auto-approve flow
- [ ] run tests - must pass before next task

### Task 6: Wire everything together in main
- [ ] connect all components in `cmd/mdrefactor/main.go`: parse args -> resolve files -> for each file: read -> call LLM -> compute diff -> show diff -> prompt -> apply
- [ ] add progress indicator: "Processing file 2/5: notes.md"
- [ ] add summary at the end: "3 files fixed, 1 skipped, 1 unchanged"
- [ ] handle graceful shutdown (Ctrl+C)
- [ ] write integration test with mock LLM server testing full flow
- [ ] run tests - must pass before next task

### Task 7: Verify acceptance criteria
- [ ] verify: single file processing works
- [ ] verify: multiple files processing works
- [ ] verify: directory recursive processing works
- [ ] verify: `--auto` flag skips approval
- [ ] verify: `--endpoint` flag configures LMStudio URL
- [ ] verify: diff display is colored and scrollable
- [ ] run full test suite
- [ ] run linter (`golangci-lint run`) - all issues must be fixed

### Task 8: [Final] Update documentation
- [ ] create README.md with: project description, installation, usage examples, flags reference
- [ ] add example usage with LMStudio setup instructions

## Technical Details

### CLI Interface
```
mdrefactor [flags] <file|dir> [file|dir...]

Flags:
  --auto         Apply all changes without asking for approval
  --endpoint     LMStudio API endpoint (default: http://localhost:1234)
```

### LMStudio API
- OpenAI-compatible: `POST {endpoint}/v1/chat/completions`
- Request body:
  ```json
  {
    "model": "local-model",
    "messages": [
      {"role": "system", "content": "Fix grammar, spelling, and syntax errors..."},
      {"role": "user", "content": "<markdown content>"}
    ],
    "temperature": 0.3
  }
  ```

### Project Structure
```
cmd/mdrefactor/main.go     - entry point, CLI parsing, orchestration
internal/llm/client.go     - LMStudio API client
internal/files/reader.go   - file discovery and reading
internal/diff/diff.go      - diff computation
internal/tui/diffview.go   - Bubble Tea diff viewer
internal/tui/approve.go    - approval prompt component
```

### Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - terminal styling
- `github.com/sergi/go-diff` - diff computation

## Post-Completion

**Manual verification:**
- Test with actual LMStudio running locally with a model loaded
- Test with various Markdown files (different sizes, languages, formatting)
- Test with large files to check LLM context limits
