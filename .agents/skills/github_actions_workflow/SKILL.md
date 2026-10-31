---
name: github_actions_workflow
description: Guidelines for creating and updating GitHub Actions workflows, ensuring explicit permissions and gh actions-lock dependency management.
---

# GitHub Actions Workflow Guidelines

## Rules & Best Practices

1. **Explicit Permissions**
   - Always specify `permissions` explicitly at the workflow or job level to enforce the principle of least privilege.

2. **Lock Actions with `gh actions-lock`**
   - Specify human-readable semantic version tags (e.g. `actions/checkout@v7.0.1`) in workflow YAML files.
   - Use `gh actions-lock` to generate and update `.github/workflows/actions.lock` to lock direct and transitive action dependencies to immutable commit SHAs.
   - Ensure `gh actions-lock --verify` passes before submitting changes.
