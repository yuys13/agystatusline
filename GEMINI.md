# agystatusline (Gemini/Antigravity CLI Developer Guide)

This file guides AI agents and developers on how to work with the `agystatusline` codebase.

## Development Commands

- **Run all unit tests**: `go test ./...`
- **Build the executable binary**: `go build -o agystatusline`
- **Run local manual integration tests**: `cat test_data.json | ./agystatusline`
- **Run interactive configuration menu (TUI)**: `./agystatusline` (without piping data)

## Development Process (TDD)

We strictly follow **Test-Driven Development (TDD)**:

1. **Red**: Write tests in `*_test.go` inside the target package before creating/modifying production code. Run tests to see them fail.
2. **Green**: Write the minimum code required in the corresponding `.go` file to pass tests.
3. **Refactor**: Refactor both production and test code for optimal design, naming, and structure.
4. **Commit**: Commit immediately after completing a refactoring step.

## Commit Guidelines

All commits must follow the **Conventional Commits** specification:
`<type>(<scope>): <subject>`

### "Why" Over "What" Rule
The subject line and body should explain **why** a change was made (its motivation or problem solved) rather than listing "what" was changed.

- **Incorrect**: `feat(cache): add git command caching logic`
- **Correct**: `feat(cache): implement persistent and in-memory Git command caching to optimize rendering speed and prevent redundant spawns`

## Architecture & Codebase Layout

- [main.go](main.go): CLI argument parsing, input TTY routing, and settings initialization.
- [types/](types/): Core telemetry and configuration structures (`StatusJSON`, `Settings`, `WidgetItem`). Used by all subpackages to prevent circular imports.
- [renderer/](renderer/): Normal and Powerline layouts renderer, ANSI escape code strip/wrap, visible text width calculations, and safe ANSI-aware truncation.
- [utils/](utils/): Git cache manager (both in-process memory and persistent file systems under `~/.cache/agystatusline/`).
- [widgets/](widgets/): Interactive widgets (Model, ContextLength, GitBranch, GitChanges).
- [tui/](tui/): Bubble Tea TUI interactive configuration menu with live statusline previews.
