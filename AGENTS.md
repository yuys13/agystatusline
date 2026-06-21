# agystatusline (AI Agents & Developer Guide)

This file guides AI agents and developers on how to work with the `agystatusline` codebase.

## Development Commands

- **Run all unit tests**: `go test ./...`
- **Build the executable binary**: `go build -o agystatusline`
- **Run local manual integration tests**: `cat test_data.json | ./agystatusline`
- **Run interactive configuration menu (TUI)**: `./agystatusline` (without piping data)
- **Format codebase**: `go fmt ./...`
- **Fix deprecated APIs**: `go fix ./...`

## TDD & Commit Guidelines

Refer to the [tdd](.agents/skills/tdd/SKILL.md) and [git_commit](.agents/skills/git_commit/SKILL.md) skills for detailed guidelines on Test-Driven Development and commit messages.

- **TDD Workflow**: Strictly follow the Red -> Green -> Refactor -> Commit loop.
- **Commit Messages**: Use Conventional Commits and explain **why** a change was made rather than "what" was changed.
- **Placeholder Domains in Tests & Docs**: Always use RFC-compliant reserved domains (such as `example.com`, `example.org`, or `example.net` according to RFC 2606) for test cases, sample codes, and documentation instead of arbitrary real-world domains.

## Architecture & Codebase Layout

- [main.go](main.go): CLI argument parsing, input TTY routing, and settings initialization.
- [types/](types/): Core telemetry and configuration structures (`StatusJSON`, `Settings`, `WidgetItem`). Used by all subpackages to prevent circular imports.
- [renderer/](renderer/): Normal and Powerline layouts renderer. Handles ANSI escape code strip/wrap, visible text width calculations, and safe ANSI-aware truncation.
- [utils/](utils/): Git cache manager (both in-process memory and persistent file systems under `~/.cache/agystatusline/`).
- [widgets/](widgets/): Interactive widgets.
- [tui/](tui/): Bubble Tea TUI interactive configuration menu with live statusline previews.

## Supported Widgets

- **Model**: Displays the active model name (`DisplayName` or `ID` from telemetry).
- **Context Length**: Displays the total input tokens (formatted in `k` or `M` if large).
- **Context Used %**: Displays the used percentage of the context window.
- **Context Remaining %**: Displays the remaining percentage of the context window.
- **Quota**: Displays quota usage and reset countdowns based on a metadata `key` (e.g., RPC limits) and custom display configurations (`quota`, `reset`, or both).
- **Git Branch**: Displays the current Git branch name (cached for performance). Supports a custom branch symbol.
- **Git Changes**: Displays insertions and deletions (e.g., `(+42,-10)`) relative to the git repository.
- **Sandbox**: Displays the `sandbox.enabled` status (e.g., `sandbox: true`).
- **Custom Text**: Displays user-defined custom static text.

## Key Features & UI/UX Designs

- **Interactive TUI Configuration**: Built with Bubble Tea. Features a live preview of the statusline at the very top of the menu layout (without border styling to mimic the real terminal display).
- **Multi-Line Editing (Edit Lines)**: Allows users to add, delete, and configure multiple statusline lines and manage their widgets through dedicated TUI menus.
- **Powerline Submenus**: Transition to dedicated submenus when selecting a Powerline theme or Powerline separator.
- **Color Level Support**: Support configuring the color output levels (ANSI 16, ANSI 256, or Truecolor) via settings and TUI.
- **Custom Caps**: Custom prefix/suffix caps (`StartCaps` and `EndCaps`) can be configured for the statusline.
- **Padding & Separator Alignment**: Non-ASCII separators automatically append a half-width space for visual alignment. Separator-level space padding is minimized; instead, widgets prepend/append spaces for clean spacing.
- **ANSI-Aware & East Asian Width Truncation**: Prevents layout corruption or color bleeding by utilizing ANSI-aware string length measurement and slice operations.
