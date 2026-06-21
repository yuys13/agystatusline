---
name: git_commit
description: Guide the agent to write why-focused commit messages using Conventional Commits and stage files individually.
---

# Git Commit Guidelines

Commit messages must be clear, concise, and structured, preceded by individual file staging. Always adhere to the following rules:

## 1. Code Formatting & Fixing
Before staging and committing Go files, always run `go fmt ./...` and `go fix ./...` to guarantee code style consistency and resolve deprecated APIs. Ensure that any resulting formatting or API fix modifications are staged along with the other changes.

## 2. Staging Files Individually
Always stage modified files explicitly and individually using `git add <file1> <file2> ...`. Do not use wildcard commands such as `git add .` or `git add -A` to avoid accidentally staging untracked, temporary, or private files (e.g., sample JSON outputs or local logs).

## 3. Format
Use the **Conventional Commits** specification:
`<type>(<scope>): <subject>`

- **Types**: `feat` (new feature), `fix` (bug fix), `refactor` (code restructuring), `test` (adding/updating tests), `chore` (maintenance, build changes), `docs` (documentation).
- **Scope**: The module, package, or component being modified (e.g., `renderer`, `cache`, `types`).

## 4. Focus on "Why" over "What"
The subject line and description should explain **why** the change was made (the motivation or problem solved) rather than merely listing the files or lines added.

- **Bad (What-focused)**:
  `feat(cache): add git command caching logic` (This just repeats what the code does)
- **Good (Why-focused)**:
  `feat(cache): implement persistent and in-memory Git command caching to optimize rendering speed and prevent redundant spawns` (Explains the motivation and benefit)

- **Bad (What-focused)**:
  `refactor(types): move types.go to types package`
- **Good (Why-focused)**:
  `refactor(types): move types definition to a separate subpackage to prevent circular dependencies in main packages`
