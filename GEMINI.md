# agystatusline (Gemini/Antigravity CLI Developer Guide)

This file guides AI agents and developers on how to work with the `agystatusline` codebase.

## Development Commands

- **Run all unit tests**: `go test ./...`
- **Build the executable binary**: `go build -o agystatusline`
- **Run local manual integration tests**: `cat test_data.json | ./agystatusline`
- **Run interactive configuration menu (TUI)**: `./agystatusline` (without piping data)

## TDD & Commit Guidelines

Refer to the [tdd_why_commit](.agents/skills/tdd_why_commit/SKILL.md) skill for detailed guidelines on Test-Driven Development and commit messages.

- **TDD Workflow**: Strictly follow the Red -> Green -> Refactor -> Commit loop.
- **Commit Messages**: Use Conventional Commits and explain **why** a change was made rather than "what" was changed.

## Architecture & Codebase Layout

- [main.go](main.go): CLI argument parsing, input TTY routing, and settings initialization.
- [types/](types/): Core telemetry and configuration structures (`StatusJSON`, `Settings`, `WidgetItem`). Used by all subpackages to prevent circular imports.
- [renderer/](renderer/): Normal and Powerline layouts renderer, ANSI escape code strip/wrap, visible text width calculations, and safe ANSI-aware truncation.
- [utils/](utils/): Git cache manager (both in-process memory and persistent file systems under `~/.cache/agystatusline/`).
- [widgets/](widgets/): Interactive widgets (Model, ContextLength, GitBranch, GitChanges).
- [tui/](tui/): Bubble Tea TUI interactive configuration menu with live statusline previews.
