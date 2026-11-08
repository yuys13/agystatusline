---
name: tdd
description: Guide the agent to use Test-Driven Development (TDD) cycles.
---

# Test-Driven Development (TDD) Workflow

This skill outlines the process for Test-Driven Development (TDD) cycles, ensuring software quality and design clarity by writing tests before implementation.

## Development Cycle (TDD)

When tasked with implementing features, follow this strict 5-step loop for every component, package, or major logical block:

1. **Red**: Write unit tests covering the expected requirements before writing any production code. Run tests to verify failure (compilation or test failure).
2. **Green**: Write the minimum necessary production code to make the tests pass.
3. **Refactor**: Clean up both the production code and the test code for readability and design without breaking functionality.
4. **Verify**: Before committing, refer to the [code_verification](../code_verification/SKILL.md) skill to run required formatting (`nix fmt`), linting (`golangci-lint`), tests, and build checks.
5. **Commit**: Git commit immediately after verification passes. Refer to the [git_commit](../git_commit/SKILL.md) skill for details on staging files and writing commit messages.

## Best Practices for Writing Go Unit Tests

When writing or refactoring Go unit tests, strictly adhere to the following principles:

1. **Table-Driven Tests**: Utilize table-driven tests (`tests := []struct{ ... }`) for parameterized inputs, edge cases, and option permutations to keep test cases structured, readable, and maintainable.
2. **Single Concern per Test Function / Subtest**: Each test function or `t.Run` subtest must focus strictly on verifying a single concern or behavior:
   - **Split Tests When Conditional Assertions Arise**: If you find yourself writing conditional branches (`if`) in test runners to decide whether to assert something or switch assertions, treat this as a clear signal that distinct concerns are mixed. Do not branch—split the test cases or subtests so that each group shares an identical expectation schema.
   - Avoid catch-all container tests (e.g. `*EdgeCases`). Decompose distinct behaviors and edge conditions into focused, dedicated test functions or subtests.
3. **Strict and Unconditional Assertions**:
   - Every test case must declare concrete, expected values (`want` vs `got`) and be asserted unconditionally. Never create "no-crash" tests without assertions to boost coverage.
   - `if` statements in test runners must strictly be used to detect failure conditions (e.g., `if got != tt.want { t.Errorf(...) }`). Never use `if` to make assertions optional or skip checks for specific cases.
   - Avoid assertion callbacks (e.g., `verify func(t *testing.T, ...)`). Declare explicit expectation fields in test structs and assert them directly in the test runner to ensure uniform verification rigor.
4. **Standard Test File Organization**: Follow standard Go conventions (`foo.go` -> `foo_test.go`). Do not create arbitrary custom test files (e.g., `stress_*.go`, `adversarial_*.go`) just for grouping test cases. Dedicated `integration_*_test.go` files are reserved for package-boundary or CLI integration tests.
