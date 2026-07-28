---
name: git_commit
description: Guide the agent to write why-focused commit messages using Conventional Commits and stage files individually.
---

# Git Commit Guidelines

Commit messages must be clear, concise, and structured, preceded by individual file staging. Always adhere to the following rules:

## 1. Code Formatting & Verification
Before staging and committing files, ensure code formatting and local quality verification steps pass (refer to the [code_verification](../code_verification/SKILL.md) skill). Ensure that any resulting formatting or API fix modifications are staged along with the other changes.

## 2. Staging Files Individually
Always stage modified files explicitly and individually using `git add <file1> <file2> ...`. Do not use wildcard commands such as `git add .` or `git add -A` to avoid accidentally staging untracked, temporary, or private files (e.g., sample JSON outputs or local logs).

## 3. Format
Use the **Conventional Commits** specification. The first line (header) must be kept concise (ideally under 50-70 characters) to avoid visual clutter in git logs. Detailed explanation, background context, and motivation must be placed in the body starting from the third line, separated from the header by a single blank line.

Wrap all body text (line 3 onwards) at approximately 72 characters per line to maintain optimal readability in standard terminal viewports.

```
<type>(<scope>): <short summary>

<detailed description, background, and motivation>
<wrapped at ~72 chars per line>
```

- **Types**: `feat` (new feature), `fix` (bug fix), `refactor` (code restructuring), `test` (adding/updating tests), `chore` (maintenance, build changes), `docs` (documentation).
- **Scope**: The module, package, or component being modified (e.g., `renderer`, `cache`, `types`).

## 4. Focus on "Why" over "What"
The commit message should explain **why** the change was made (the motivation or problem solved) rather than merely listing the files or lines added. Put the short reason in the first line and expand on the details in the body, ensuring lines in the body are wrapped at around 72 characters.

- **Bad (Too long header / What-focused)**:
  `feat(cache): implement persistent and in-memory Git command caching to optimize rendering speed and prevent redundant spawns`
- **Good**:
  ```
  feat(cache): optimize Git command performance with caching

  Introduce both in-memory and persistent file-system caching for Git status
  to prevent redundant process spawns and improve CLI rendering speed.
  ```

- **Bad (What-focused)**:
  `refactor(types): move types.go to types package`
- **Good**:
  ```
  refactor(types): resolve circular dependencies in main packages

  Move core telemetry and configuration structure definitions to a separate
  `types` subpackage to prevent circular package imports.
  ```

