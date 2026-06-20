---
name: persistent_state_caching
description: Guide the agent to design fast, reliable process cache mechanisms combining in-memory and atomic disk persistence to speed up recurrent CLI executions.
---

# Persistent State Caching Guidelines for CLI Tools

CLI tools that run frequently (e.g., statuslines rendering every 300ms) must avoid spawning heavy external commands (like `git status`, `git diff`) on every execution to prevent terminal lag and high CPU usage. This guide defines a robust "Memory + Disk" caching pattern with dynamic invalidation.

## Architecture

Implement a two-layered caching system:
1. **In-Memory Cache (In-Process)**: Extremely fast; scoped to the current execution thread or process.
2. **Persistent Disk Cache**: Shared across concurrent processes (e.g., multiple statusline instances running in split panes). Located in a standard directory (e.g., `~/.cache/<app-name>/`).

## Implementation Principles

### 1. Key Invalidation Triggered by File Modification Time (mtime)
Do not rely solely on TTL (Time-To-Live) expiration. Check file modifications:
- For Git queries, monitor the modification times (`mtime`) of `.git/HEAD` and `.git/index`.
- If the current file `mtime` values differ from the cache record's values, invalidate the cache immediately.
- If they match, allow caching up to the configured TTL limit (e.g., 5 seconds).

### 2. Context-Sensitive Cache Keys
Cache keys must incorporate context to prevent returning incorrect data for a different environment:
- Always append the resolved current working directory (CWD) to the cache key (e.g. `cmd + "|" + CWD`).

### 3. Atomic Disk Writes
To avoid corrupting cache files when multiple CLI instances write concurrently:
- **Never write directly to the target file.**
- Create a temporary file in the same directory (e.g., `<target>.pid.tmp`).
- Write the JSON serialized cache to the temporary file.
- Close the file, then atomically rename it to the target file (`os.Rename` or equivalent). This guarantees that concurrent readers see either a fully written old cache or a fully written new cache, preventing partial/torn writes.

### 4. Graceful Fallbacks
Caching is a non-critical optimization. If a disk cache read or write fails due to permission errors or locking issues, catch the error and fall back gracefully to running the command directly without crashing the app.
