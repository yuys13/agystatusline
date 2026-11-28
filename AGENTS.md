# agystatusline (AI Agents & Developer Guide)

This file guides AI agents and developers on how to work with the `agystatusline` codebase.

## Development Commands

- **Run all unit tests**: `go test -race -shuffle=on ./...`
- **Build the executable binary**: `go build -o agystatusline`
- **Run local manual integration tests**: `cat tests/data/payload_sample.json | go run .`
- **Run interactive configuration menu (TUI)**: `go run .` (without piping data)
- **Format codebase**: `go fmt ./...`
- **Fix deprecated APIs**: `go fix ./...`

## Workflow & Guardrails

- **Skills**: Follow [.agents/skills/tdd/SKILL.md](.agents/skills/tdd/SKILL.md), [.agents/skills/code_verification/SKILL.md](.agents/skills/code_verification/SKILL.md), [.agents/skills/git_commit/SKILL.md](.agents/skills/git_commit/SKILL.md), [.agents/skills/github_pr_workflow/SKILL.md](.agents/skills/github_pr_workflow/SKILL.md), and [.agents/skills/ansi_aware_string_manipulation/SKILL.md](.agents/skills/ansi_aware_string_manipulation/SKILL.md).
- **Placeholder Domains in Tests & Docs**: Follow RFC 2606 / RFC 6761 instead of arbitrary real-world domains. Use `example.com`, `example.org`, `example.net`, or `.example` strictly for documentation, sample data, and offline/mocked test fixtures—never make live network requests to them in tests. Use local mocks (`httptest` / `localhost`) for live test requests, `.test` for local test environments, and `.invalid` for guaranteed-invalid domains.
- **UI/UX & Rendering Invariants**:
  - **TUI Live Preview**: Rendered at the very top of the menu layout without border styling to mimic the real terminal display, with dedicated submenus for multi-line editing, Powerline themes, and separators.
  - **Color & Caps**: Supports ANSI 16, ANSI 256, and Truecolor (`color_level` 1–3) plus custom `StartCaps` and `EndCaps`.
  - **Padding & Separator Alignment**: Non-ASCII separators automatically append a half-width space for visual alignment. Keep separator-level space padding minimal; widgets manage their own leading/trailing spaces.
- **Documentation Maintenance**: Keep `AGENTS.md` concise without inline specification tables or schema examples. Whenever widgets or configuration options change, update [docs/widgets.md](docs/widgets.md) and [docs/configuration.md](docs/configuration.md) accordingly.

## Architecture & Codebase Layout

- [main.go](main.go): CLI entrypoint, TOML settings initialization (`~/.config/agystatusline/settings.toml`), input TTY routing, and statusline rendering loop.
- [types/](types/): Core telemetry and configuration structures (`StatusJSON`, `Settings`, `PowerlineConfig`, `GeneralConfig`, `WidgetItem`). Implements TOML unmarshaling/marshaling supporting both string shorthand and inline tables. Used across subpackages to prevent circular imports.
- [renderer/](renderer/): Standard and Powerline layout renderer. Handles ANSI escape code strip/wrap, visible text width calculations, multi-color palette mappings, gradient text interpolation, and safe ANSI-aware truncation.
- [utils/](utils/): Git cache manager (in-memory + atomic file persistence under `~/.cache/agystatusline/`) and terminal utility helpers.
- [widgets/](widgets/): 14 core widget categories (26 widget identifiers) implementing the `Widget` interface.
- [tui/](tui/): Bubble Tea TUI interactive configuration menu with live statusline previews and multi-line editor.

## Detailed References

Read these documents on demand rather than keeping them inline:

- [docs/widgets.md](docs/widgets.md): Supported widgets (14 categories / 26 identifiers), default keys/labels, example outputs, and dynamic color rules. Read when adding or modifying widgets.
- [docs/configuration.md](docs/configuration.md): TOML configuration schema, section options, and widget item syntax (shorthand strings & inline tables). Read when changing settings schema, TOML parsing, or TUI config options.
