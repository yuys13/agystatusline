# Configuration Reference

The configuration file is stored in TOML format at `~/.config/agystatusline/settings.toml` (or custom path via `--config <path>`).

## Top-Level Structure

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

# Telemetry / state payload persistence (optional)
[payload_dump]
enabled = false
path = "~/.local/share/agystatusline/payload.json"

# Quota ordering, exclusion, and custom label mappings (optional)
[quota]
order = ["gemini-5h", "gemini-weekly", "gemini-argon-wtu-weekly", "3p-5h", "3p-weekly"]
exclude = []

[quota.labels]
gemini-5h = "Gemini 5h"
gemini-weekly = "Gemini 7d"
gemini-argon-wtu-weekly = "Argon 7d"
3p-5h = "Claude 5h"
3p-weekly = "Claude 7d"
```

## Configuration Sections

### `[powerline]`

- **`enabled`** (`bool`, default: `false`): Enables Powerline background coloring and glyph separators.
- **`theme`** (`string`, default: `"nord-aurora"`): Color theme name. Supported themes:
  - `"nord"`, `"nord-aurora"`, `"monokai"`, `"solarized"`, `"minimal"`, `"dracula"`, `"catppuccin"`, `"gruvbox"`, `"onedark"`, `"tokyonight"`
- **`separator`** (`string`, default: `"\uE0B0"`): Glyphs placed between widgets (e.g., Arrow `\uE0B0`, Round `\uE0B4`, Flame `\uE0C0`, Hexagon `\uE0C6`, Slanted `\uE0C8`).
- **`start_caps`** (`string`, default: `""`): Leading glyph at the beginning of each statusline line (e.g., `\uE0B2`, `\uE0B6`).
- **`end_caps`** (`string`, default: `""`): Trailing glyph at the end of each statusline line (e.g., `\uE0B0`, `\uE0B4`).

### `[general]`

- **`color_level`** (`int`, default: `1`): Terminal color mode:
  - `1`: ANSI 16 colors (Standard)
  - `2`: ANSI 256 colors
  - `3`: Truecolor (24-bit RGB)
- **`git_cache_ttl`** (`int`, default: `5`): Cache time-to-live in seconds for git command queries.
- **`separator`** (`string`, default: `" · "`): Widget separator string in standard mode (Powerline disabled).
- **`padding`** (`string`, default: `""`): Optional padding text added around widgets.
- **`minimalist`** (`bool`, default: `false`): When `true`, suppresses widget titles/labels and displays only values.

### `[payload_dump]`

- **`enabled`** (`bool`, default: `false`): When `true`, saves the verbatim JSON state payload streamed on stdin to the file path specified in `path`.
- **`path`** (`string`, default: `"~/.local/share/agystatusline/payload.json"`): Destination file path for persisting the telemetry JSON (falls back to default path if empty or omitted). Uses atomic replace and sets `0600` file permissions.

### `[quota]` & `[quota.labels]`

- **`order`** (`[]string`, default: `["gemini-5h", "gemini-weekly", "gemini-argon-wtu-weekly", "3p-5h", "3p-weekly"]`): Display order of quota buckets when rendering `quota-all` or `quota-bar-all`. When specified in `settings.toml`, overrides the default order list; any remaining non-excluded keys present in telemetry are appended in alphabetical order.
- **`exclude`** (`[]string`, default: `[]`): Quota keys to exclude when rendering `quota-all` or `quota-bar-all` (e.g. `["3p-5h", "3p-weekly"]`).
- **`[quota.labels].<quota-key>`** (`string`): Overrides the display label for the given telemetry quota key across all quota and quota-bar widgets (unless overridden per-widget via `text`). Built-in defaults:
  - `gemini-5h` = `"Gemini 5h"`
  - `gemini-weekly` = `"Gemini 7d"`
  - `gemini-argon-wtu-weekly` = `"Argon 7d"`
  - `3p-5h` = `"Claude 5h"`
  - `3p-weekly` = `"Claude 7d"`

## Widget Item Syntax

Widgets inside `lines` can be defined using either **shorthand strings** or **inline tables**:

1. **Shorthand String Syntax**:
   - Plain identifier: `"model"`, `"agent-state"`, `"context-bar"`, `"artifacts"`, `"subagents"`, `"tasks"`, `"sandbox"`, `"git-repo"`, `"git-branch"`, `"git-ahead-behind"`, `"git-changes"`
   - Pre-bound / all-buckets quota: `"quota-all"`, `"quota-5h"`, `"quota-7d"`, `"quota-argon-7d"`, `"quota-3p-5h"`, `"quota-3p-7d"`, `"quota-bar-all"`, `"quota-bar-5h"`, `"quota-bar-7d"`, `"quota-bar-argon-7d"`, `"quota-bar-3p-5h"`, `"quota-bar-3p-7d"`
   - Parameter shorthand with `:`:
     - `"quota:gemini-5h"` (sets `key`)
     - `"quota-bar:gemini-weekly"` (sets `key`)
     - `"custom-text:PROD"` or `"custom:PROD"` (sets `text`)
     - `"git-repo:📦 "` (sets `symbol`)
     - `"git-branch:🌿 "` (sets `symbol`)
     - `"git-ahead-behind:⇅ "` (sets `symbol`)

2. **Inline Table Syntax**:
   Allows fine-grained customization of individual widgets:
   - `{ type = "quota-5h", color = "brightCyan" }`
   - `{ type = "quota", key = "rpc", text = "RPC", color = "brightGreen", raw = true }`
   - `{ type = "git-repo", symbol = "📦 ", color = "brightBlue" }`
   - `{ type = "git-branch", symbol = "⎇ ", color = "brightMagenta" }`
   - `{ type = "git-ahead-behind", symbol = "⇅ ", color = "brightCyan" }`
   - `{ type = "custom-text", text = "PROD", color = "brightRed", raw = false }`
   - `{ type = "agent-state", hint = "esc to cancel" }` (shows `⚙ WORKING (esc to cancel)` while the agent is busy)

   **Supported Table Properties**:
   - `type` (`string`, required): Widget identifier name.
   - `key` (`string`, optional): Quota lookup key in telemetry (e.g., `"gemini-5h"`, `"gemini-weekly"`, `"3p-5h"`, `"3p-weekly"`).
   - `text` (`string`, optional): Custom label or text content.
   - `symbol` (`string`, optional): Custom icon/symbol prefix (e.g. for `git-repo`, `git-branch`, `git-ahead-behind`).
   - `color` (`string`, optional): Widget color override. Supported formats:
     - Named color: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `brightBlack`, `brightRed`, `brightGreen`, `brightYellow`, `brightBlue`, `brightMagenta`, `brightCyan`, `brightWhite`
     - ANSI 256: `ansi256:<0-255>` (e.g., `ansi256:208`)
     - Truecolor Hex: `hex:<RRGGBB>` (e.g., `hex:FF5733`)
     - Gradient: `gradient:hex:FF0000,hex:0000FF` or `gradient:red,blue`
   - `raw` (`bool`, optional): When `true`, suppresses the widget title label and outputs only the raw value.
   - `hint` (`string`, optional): Text appended in parentheses while the widget has an active state. Currently rendered by `agent-state` for any state other than `idle`, for example to advertise the interrupt key: `{ type = "agent-state", hint = "esc to cancel" }`. Unknown busy states also show it.
