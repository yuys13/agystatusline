---
name: tdd_why_commit
description: Guide the agent to use Test-Driven Development (TDD) and commit with a focus on "why" rather than "what" using Conventional Commits.
---

# TDD & "Why" Focused Commit Workflow

This skill outlines the process for Test-Driven Development (TDD) cycles integrated with micro-commits using Conventional Commits, prioritizing the business logic or design rationale ("why") over raw changes ("what").

## Development Cycle (TDD)

When tasked with implementing features, follow this strict 4-step loop for every component, package, or major logical block:

1. **Red**: Write unit tests (`*_test.go` or equivalent for other languages) covering the expected requirements before writing any production code. Run tests to verify failure (compilation or test failure).
2. **Green**: Write the minimum necessary production code to make the tests pass.
3. **Refactor**: Clean up both the production code and the test code for readability and design without breaking functionality.
4. **Commit**: Git commit immediately after a successful refactoring step.

## Git Commit Guidelines

Commit messages must be clear, concise, and structured. Always adhere to the following rules:

### 1. Format
Use the **Conventional Commits** specification:
`<type>(<scope>): <subject>`

- **Types**: `feat` (new feature), `fix` (bug fix), `refactor` (code restructuring), `test` (adding/updating tests), `chore` (maintenance, build changes), `docs` (documentation).
- **Scope**: The module, package, or component being modified (e.g., `renderer`, `cache`, `types`).

### 2. Focus on "Why" over "What"
The subject line and description should explain **why** the change was made (the motivation or problem solved) rather than merely listing the files or lines added.

- **Bad (What-focused)**:
  `feat(cache): add git command caching logic` (This just repeats what the code does)
- **Good (Why-focused)**:
  `feat(cache): implement persistent and in-memory Git command caching to optimize rendering speed and prevent redundant spawns` (Explains the motivation and benefit)

- **Bad (What-focused)**:
  `refactor(types): move types.go to types package`
- **Good (Why-focused)**:
  `refactor(types): move types definition to a separate subpackage to prevent circular dependencies in main packages`

### 3. Staging Files Individually
Always stage modified files explicitly and individually using `git add <file1> <file2> ...`. Do not use wildcard commands such as `git add .` or `git add -A` to avoid accidentally staging untracked, temporary, or private files (e.g., sample JSON outputs or local logs).

