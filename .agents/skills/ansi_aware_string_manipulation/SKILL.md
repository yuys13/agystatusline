---
name: ansi_aware_string_manipulation
description: Prevent layout corruption and styling leaks in CLI/TUI apps by ensuring all string operations (length, slice, truncate) are aware of ANSI escape sequences and East Asian Width.
---

# ANSI-Aware String Manipulation Guidelines

When developing command-line interfaces (CLI) or text user interfaces (TUI) that output styled text, using standard string operations (like slicing `str[:limit]` or counting bytes `len(str)`) is a major anti-pattern. This guide ensures that styling and layout remain intact by correctly handling ANSI escape codes and character display widths.

## The Problem

Standard string operations do not distinguish between printable characters and invisible escape sequences:
1. **Length Errors**: `len("\x1b[31mHello\x1b[0m")` returns 14 instead of 5, causing misalignment in padded layouts.
2. **Style Leaks**: Slicing in the middle of a string (e.g. `str[:limit]`) can cut off the SGR reset code `\x1b[0m`, causing the terminal color to leak into the rest of the shell.
3. **Broken Sequences**: Slicing inside an escape sequence (e.g., inside `\x1b[31m`) results in corrupt control codes, printing garbled text to the terminal.
4. **Multibyte Character Width**: Characters like Japanese, Chinese, Korean, or emojis occupy 2 terminal cells, but have a rune length of 1 and byte length > 1.

## Rules for Safe Operations

### 1. Counting Visible Width
Never use raw string length. To determine alignment or padding sizes:
- First, strip all ANSI SGR (styles) and OSC (hyperlinks) sequences.
- Measure the printable width of the remaining characters using a dedicated library (e.g., `go-runewidth` in Go, `string-width` in JS) to account for full-width characters (East Asian Width).

### 2. ANSI-Safe Truncation
When truncating a styled string to fit a terminal width limit:
- Iterate through the string character-by-character (runes) or token-by-token.
- Parse and bypass entire ANSI CSI (`\x1b[...]`) and OSC (`\x1b]...`) sequences, adding them directly to the output buffer without counting toward the display width limit.
- Only increment the display width for printable characters. Stop when the limit is reached.
- **Always close opened states**: If an OSC 8 hyperlink or an SGR decoration is left open at the truncation point, append the corresponding reset sequences (e.g., `\x1b]8;;\x1b\\` or `\x1b[0m`) before appending an ellipsis (`...`).

### 3. Separation of Styles and Layout
- Apply padding and separators *after* truncating or aligning individual widget outputs.
- Keep style-rendering utilities (such as color wrappers) separate from string layout calculations.
