---
name: code_verification
description: Guidelines for code formatting, linting, and local quality verification before committing changes.
---

# Code Formatting & Verification Guidelines

Before staging and committing changes, always ensure code formatting, static analysis, and local build/test checks are completely satisfied:

## 1. Formatting
- Run `nix fmt` (or `go fmt ./...`) to guarantee code and configuration style consistency.
- Ensure any formatting modifications are staged alongside the relevant code changes.

## 2. Verification & Testing
- Run `go test ./...` to ensure all unit tests pass locally.
- Run `golangci-lint run` (or `nix develop --command golangci-lint run`) to verify static code analysis.
- Run `nix flake check --all-systems` to verify Nix build rules, treefmt checks, and cross-platform compatibility.
- Ensure `go build -o agystatusline` builds successfully without warnings or errors.
