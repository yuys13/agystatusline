---
name: code_verification
description: Guidelines for code formatting, linting, and local quality verification before committing changes.
---

# Code Formatting & Verification Guidelines

Before staging and committing changes, run only the checks scoped to the modified files. Chain commands with `&&` in a single shell invocation and avoid verbose flags (such as `go test -v`) unless diagnosing a failure, to minimize turn count and output token usage.

## 1. Scope by Change Type

- **Docs / Config only (`.md`, `.yml`, `.yaml`, etc.)**:
  - Run `nix fmt` (when `flake.nix` is present).
- **Go code changes (before each commit)**:
  - Run in a single chained command:
    `nix fmt && go test -race -shuffle=on ./... && golangci-lint run`
  - Without Nix (no `flake.nix` or `nix` unavailable), use `go fmt ./...` instead of `nix fmt`.
  - If `go` or `golangci-lint` is not on `PATH`, run them via `nix develop --command ...`.
- **Nix / dependency changes (`flake.nix`, `go.mod`, `go.sum`) or before opening a PR**:
  - Also run `nix flake check --no-build --all-systems` (matching CI) to verify Nix evaluation, `treefmt` checks, and cross-platform definitions without duplicating Go builds/tests.

## 2. Notes

- Always stage any formatting changes produced by `nix fmt` / `go fmt` (`git add`) before committing.
