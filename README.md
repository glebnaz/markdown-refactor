# mdrefactor

CLI tool that fixes grammar, spelling, and syntax errors in Markdown files using a local LLM (LMStudio).

Shows a colored diff of proposed changes and asks for your approval before applying them.

## Features

- Fix grammar, spelling, and syntax in Markdown without changing meaning
- Colored diff display for reviewing changes
- Interactive approve/skip/quit workflow per file
- `--auto` mode for batch processing without prompts
- Supports single files, multiple files, and directories (recursive `*.md` discovery)
- Uses local LMStudio (OpenAI-compatible API) - your data stays on your machine

## Prerequisites

- Go 1.21+
- [LMStudio](https://lmstudio.ai/) running locally with a model loaded

## Installation

```bash
go install github.com/glebnaz/markdown-refactor/cmd/mdrefactor@latest
```

Or build from source:

```bash
git clone https://github.com/glebnaz/markdown-refactor.git
cd markdown-refactor
go build -o mdrefactor ./cmd/mdrefactor
```

## LMStudio Setup

1. Download and install [LMStudio](https://lmstudio.ai/)
2. Download a model (any instruction-following model works, e.g. Llama 3, Mistral, Qwen)
3. Load the model in LMStudio
4. Start the local server - by default it runs on `http://localhost:1234`
5. Run `mdrefactor` against your Markdown files

## Usage

```
mdrefactor [flags] <file|dir> [file|dir...]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--auto` | `false` | Apply all changes without asking for approval |
| `--endpoint` | `http://localhost:1234` | LMStudio API endpoint |

### Examples

Fix a single file (interactive):

```bash
mdrefactor README.md
```

Fix multiple files:

```bash
mdrefactor docs/guide.md docs/api.md
```

Fix all Markdown files in a directory recursively:

```bash
mdrefactor docs/
```

Auto-apply all changes without prompts:

```bash
mdrefactor --auto docs/
```

Use a custom LMStudio endpoint:

```bash
mdrefactor --endpoint http://192.168.1.100:1234 README.md
```

### Interactive Mode

For each file with changes, mdrefactor will:

1. Show a colored diff (green = additions, red = removals)
2. Prompt you to approve:
   - **Y** - Apply changes to the file
   - **N** - Skip this file
   - **Q** - Quit processing

After all files are processed, a summary is displayed:

```
3 file(s) fixed, 1 skipped, 1 unchanged
```

## Project Structure

```
cmd/mdrefactor/main.go       - Entry point, CLI parsing, orchestration
internal/llm/client.go       - LMStudio API client
internal/files/reader.go     - File discovery and reading
internal/files/writer.go     - File writing
internal/diff/diff.go        - Diff computation
internal/tui/diffview.go     - Colored diff viewer (Bubble Tea)
internal/tui/approve.go      - Approval prompt (Bubble Tea)
```

## License

MIT
