# agystatusline (AI Agents & Developer Guide)

This file guides AI agents and developers on how to work with the `agystatusline` codebase.

## Development Commands

- **Run all unit tests**: `go test -race -shuffle=on ./...`
- **Build the executable binary**: `go build -o agystatusline`
- **Run local manual integration tests**: `cat test_data.json | ./agystatusline`
- **Run interactive configuration menu (TUI)**: `./agystatusline` (without piping data)
- **Format codebase**: `go fmt ./...`
- **Fix deprecated APIs**: `go fix ./...`

## TDD & Commit Guidelines

Refer to the [tdd](.agents/skills/tdd/SKILL.md), [git_commit](.agents/skills/git_commit/SKILL.md), [code_verification](.agents/skills/code_verification/SKILL.md), and [github_pr_workflow](.agents/skills/github_pr_workflow/SKILL.md) skills for detailed guidelines on Test-Driven Development, formatting, verification, commit messages, and PR workflows.

- **TDD Workflow**: Strictly follow the Red -> Green -> Refactor -> Verify -> Commit loop.
- **Standard Test File Organization**: Follow standard Go conventions (`foo.go` -> `foo_test.go`). Do not create arbitrary custom test files (e.g., `stress_*.go`, `adversarial_*.go`) just for grouping test cases.
- **Test Quality Principles**: Always use table-driven tests (`tests := []struct{...}`) for parameterized scenarios, restrict each test/subtest to a single concern, and ensure all tests strictly assert expected outputs (do not create dummy tests without assertions just for coverage).
- **Code Verification**: Before committing or finishing any code modifications, always view and execute the checks in [code_verification](.agents/skills/code_verification/SKILL.md) (`nix fmt`, `golangci-lint`, `go test -race`, `go build`).
- **Commit Messages**: Use Conventional Commits and explain **why** a change was made rather than "what" was changed.
- **Placeholder Domains in Tests & Docs**: Always use RFC-compliant reserved domains (such as `example.com`, `example.org`, or `example.net` according to RFC 2606) for test cases, sample codes, and documentation instead of arbitrary real-world domains.

## Architecture & Codebase Layout

- [main.go](main.go): CLI entrypoint, TOML settings initialization (`~/.config/agystatusline/settings.toml`), input TTY routing, and statusline rendering loop.
- [types/](types/): Core telemetry and configuration structures (`StatusJSON`, `Settings`, `PowerlineConfig`, `GeneralConfig`, `WidgetItem`). Implements TOML unmarshaling/marshaling supporting both string shorthand and inline tables. Used across subpackages to prevent circular imports.
- [renderer/](renderer/): Standard and Powerline layout renderer. Handles ANSI escape code strip/wrap, visible text width calculations, multi-color palette mappings, gradient text interpolation, and safe ANSI-aware truncation.
- [utils/](utils/): Git cache manager (in-memory + atomic file persistence under `~/.cache/agystatusline/`) and terminal utility helpers.
- [widgets/](widgets/): 12 core widget categories (20 widget identifiers) implementing the `Widget` interface.
- [tui/](tui/): Bubble Tea TUI interactive configuration menu with live statusline previews and multi-line editor.

## Supported Widgets

`agystatusline` supports **12 core widget categories** across **20 widget identifiers**:

| #   | Core Category    | Widget Identifiers (`type`)                                                               | Default Key / Label                                                                                                                                                                             | Example Output                                                           | Description & Color Dynamics                                                                                                                          |
| --- | ---------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Agent State**  | `agent-state`                                                                             | -                                                                                                                                                                                               | `● READY`, `◆ THINKING`, `⚙ WORKING`, `🔧 TOOL`                          | Displays current agent status. Colors: idle (`brightGreen`), thinking (`brightYellow`), working (`brightCyan`), tool_use (`brightMagenta`).           |
| 2   | **Model**        | `model`                                                                                   | -                                                                                                                                                                                               | `Gemini 2.5 Pro (High)`                                                  | Displays the active model name (`DisplayName` or `ID` fallback). Color: `brightMagenta`.                                                              |
| 3   | **Context Bar**  | `context-bar`                                                                             | `ctx`                                                                                                                                                                                           | `ctx ███············ 20.0%`                                              | Visual progress bar of context window usage. Color dynamically shifts: `>=90%` (`brightRed`), `>=60%` (`brightYellow`), default (`brightWhite`).      |
| 4   | **Artifacts**    | `artifacts`                                                                               | `artifacts`                                                                                                                                                                                     | `artifacts 3`                                                            | Displays count of generated session artifacts. Color: `brightWhite`.                                                                                  |
| 5   | **Subagents**    | `subagents`                                                                               | `subagents`                                                                                                                                                                                     | `subagents 2`                                                            | Displays count of active subagents. Color: `brightWhite`.                                                                                             |
| 6   | **Tasks**        | `tasks`                                                                                   | `tasks`                                                                                                                                                                                         | `tasks 1`                                                                | Displays count of background tasks. Color: `brightWhite`.                                                                                             |
| 7   | **Sandbox**      | `sandbox`                                                                                 | `sandbox`                                                                                                                                                                                       | `sandbox on` / `sandbox off`                                             | Displays sandbox execution status. Colors: enabled (`brightGreen`), disabled (`brightBlack`).                                                         |
| 8   | **Git Branch**   | `git-branch`                                                                              | `⎇ `                                                                                                                                                                                            | `⎇ main*` / `⎇ no git`                                                   | Displays Git branch name with dirty marker `*`. Supports custom prefix symbol. Colors: dirty (`brightRed`), clean (`brightBlue`).                     |
| 9   | **Git Changes**  | `git-changes`                                                                             | -                                                                                                                                                                                               | `(+42,-10)` / `(no git)`                                                 | Displays Git repository insertions and deletions. Color: `yellow`.                                                                                    |
| 10  | **Quota (Text)** | `quota`<br>`quota-5h`<br>`quota-7d`<br>`quota-3p-5h`<br>`quota-3p-7d`                     | `quota`: (explicit key)<br>`quota-5h`: `gemini-5h` ("5h")<br>`quota-7d`: `gemini-weekly` ("7d")<br>`quota-3p-5h`: `3p-5h` ("3p-5h")<br>`quota-3p-7d`: `3p-weekly` ("3p-7d")                     | `5h 50.19% (2h 28m)`<br>`7d 90.91% (6d 13h)`<br>`3p-5h 100.00% (4h 59m)` | Displays quota percentage and reset countdown. Color: `brightWhite`.                                                                                  |
| 11  | **Quota Bar**    | `quota-bar`<br>`quota-bar-5h`<br>`quota-bar-7d`<br>`quota-bar-3p-5h`<br>`quota-bar-3p-7d` | `quota-bar`: (explicit key)<br>`quota-bar-5h`: `gemini-5h` ("5h")<br>`quota-bar-7d`: `gemini-weekly` ("7d")<br>`quota-bar-3p-5h`: `3p-5h` ("3p-5h")<br>`quota-bar-3p-7d`: `3p-weekly` ("3p-7d") | `5h █████····· 50.2% (2h 28m)`<br>`7d █████████· 90.9% (6d 13h)`         | Graphical progress bar of quota with percentage and reset countdown. Colors: `>=50%` (`brightGreen`), `>=20%` (`brightYellow`), `<20%` (`brightRed`). |
| 12  | **Custom Text**  | `custom-text` (alias: `custom`)                                                           | -                                                                                                                                                                                               | User-defined static text                                                 | Renders custom string specified via `text` property or colon shorthand (e.g. `custom-text:PROD`). Color: `white` or custom.                           |

## TOML Configuration Schema & Examples

The configuration file is stored in TOML format at `~/.config/agystatusline/settings.toml` (or custom path via `--config <path>`).

### Top-Level Structure

```toml
# Array of lines, where each line is an array of widget items
lines = [
  [
    "agent-state",
    "model",
    "context-bar",
    "artifacts",
    "subagents",
    "tasks",
    "sandbox"
  ],
  [
    "git-branch",
    "git-changes"
  ],
  [
    "quota-bar-5h",
    "quota-bar-7d"
  ]
]

# Powerline styling configuration
[powerline]
enabled = false
theme = "nord-aurora"
separator = "\uE0B0"
start_caps = ""
end_caps = ""

# General display and runtime behavior
[general]
color_level = 1
git_cache_ttl = 5
separator = " · "
padding = ""
minimalist = false
```

### Configuration Sections

#### `[powerline]`

- **`enabled`** (`bool`, default: `false`): Enables Powerline background coloring and glyph separators.
- **`theme`** (`string`, default: `"nord-aurora"`): Color theme name. Supported themes:
  - `"nord"`, `"nord-aurora"`, `"monokai"`, `"solarized"`, `"minimal"`, `"dracula"`, `"catppuccin"`, `"gruvbox"`, `"onedark"`, `"tokyonight"`
- **`separator`** (`string`, default: `"\uE0B0"`): Glyphs placed between widgets (e.g., Arrow `\uE0B0`, Round `\uE0B4`, Flame `\uE0C0`, Hexagon `\uE0C6`, Slanted `\uE0C8`).
- **`start_caps`** (`string`, default: `""`): Leading glyph at the beginning of each statusline line (e.g., `\uE0B2`, `\uE0B6`).
- **`end_caps`** (`string`, default: `""`): Trailing glyph at the end of each statusline line (e.g., `\uE0B0`, `\uE0B4`).

#### `[general]`

- **`color_level`** (`int`, default: `1`): Terminal color mode:
  - `1`: ANSI 16 colors (Standard)
  - `2`: ANSI 256 colors
  - `3`: Truecolor (24-bit RGB)
- **`git_cache_ttl`** (`int`, default: `5`): Cache time-to-live in seconds for git command queries.
- **`separator`** (`string`, default: `" · "`): Widget separator string in standard mode (Powerline disabled).
- **`padding`** (`string`, default: `""`): Optional padding text added around widgets.
- **`minimalist`** (`bool`, default: `false`): When `true`, suppresses widget titles/labels and displays only values.

### Widget Item Syntax

Widgets inside `lines` can be defined using either **shorthand strings** or **inline tables**:

1. **Shorthand String Syntax**:
   - Plain identifier: `"model"`, `"agent-state"`, `"context-bar"`, `"artifacts"`, `"subagents"`, `"tasks"`, `"sandbox"`, `"git-branch"`, `"git-changes"`
   - Pre-bound quota: `"quota-5h"`, `"quota-7d"`, `"quota-3p-5h"`, `"quota-3p-7d"`, `"quota-bar-5h"`, `"quota-bar-7d"`, `"quota-bar-3p-5h"`, `"quota-bar-3p-7d"`
   - Parameter shorthand with `:`:
     - `"quota:gemini-5h"` (sets `key`)
     - `"quota-bar:gemini-weekly"` (sets `key`)
     - `"custom-text:PROD"` or `"custom:PROD"` (sets `text`)
     - `"git-branch:🌿 "` (sets `symbol`)

2. **Inline Table Syntax**:
   Allows fine-grained customization of individual widgets:
   - `{ type = "quota-5h", color = "brightCyan" }`
   - `{ type = "quota", key = "rpc", text = "RPC", color = "brightGreen", raw = true }`
   - `{ type = "git-branch", symbol = "⎇ ", color = "brightMagenta" }`
   - `{ type = "custom-text", text = "PROD", color = "brightRed", raw = false }`

   **Supported Table Properties**:
   - `type` (`string`, required): Widget identifier name.
   - `key` (`string`, optional): Quota lookup key in telemetry (e.g., `"gemini-5h"`, `"gemini-weekly"`, `"3p-5h"`, `"3p-weekly"`).
   - `text` (`string`, optional): Custom label or text content.
   - `symbol` (`string`, optional): Custom icon/symbol prefix (e.g. for `git-branch`).
   - `color` (`string`, optional): Widget color override. Supported formats:
     - Named color: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `brightBlack`, `brightRed`, `brightGreen`, `brightYellow`, `brightBlue`, `brightMagenta`, `brightCyan`, `brightWhite`
     - ANSI 256: `ansi256:<0-255>` (e.g., `ansi256:208`)
     - Truecolor Hex: `hex:<RRGGBB>` (e.g., `hex:FF5733`)
     - Gradient: `gradient:hex:FF0000,hex:0000FF` or `gradient:red,blue`
   - `raw` (`bool`, optional): When `true`, suppresses the widget title label and outputs only the raw value.

## Key Features & UI/UX Designs

- **Interactive TUI Configuration**: Built with Bubble Tea. Features a live preview of the statusline at the very top of the menu layout (without border styling to mimic the real terminal display).
- **Multi-Line Editing (Edit Lines)**: Allows users to add, delete, and configure multiple statusline lines and manage their widgets through dedicated TUI menus.
- **Powerline Submenus**: Transition to dedicated submenus when selecting a Powerline theme or Powerline separator.
- **Color Level Support**: Support configuring the color output levels (ANSI 16, ANSI 256, or Truecolor) via settings and TUI.
- **Custom Caps**: Custom prefix/suffix caps (`StartCaps` and `EndCaps`) can be configured for the statusline.
- **Padding & Separator Alignment**: Non-ASCII separators automatically append a half-width space for visual alignment. Separator-level space padding is minimized; instead, widgets prepend/append spaces for clean spacing.
- **ANSI-Aware & East Asian Width Truncation**: Prevents layout corruption or color bleeding by utilizing ANSI-aware string length measurement and slice operations.
