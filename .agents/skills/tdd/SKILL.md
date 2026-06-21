---
name: tdd
description: Guide the agent to use Test-Driven Development (TDD) cycles.
---

# Test-Driven Development (TDD) Workflow

This skill outlines the process for Test-Driven Development (TDD) cycles, ensuring software quality and design clarity by writing tests before implementation.

## Development Cycle (TDD)

When tasked with implementing features, follow this strict 4-step loop for every component, package, or major logical block:

1. **Red**: Write unit tests (`*_test.go` or equivalent for other languages) covering the expected requirements before writing any production code. Run tests to verify failure (compilation or test failure).
2. **Green**: Write the minimum necessary production code to make the tests pass.
3. **Refactor**: Clean up both the production code and the test code for readability and design without breaking functionality.
4. **Commit**: Git commit immediately after a successful refactoring step. Refer to the [git_commit](../git_commit/SKILL.md) skill for details on staging files and writing commit messages.
