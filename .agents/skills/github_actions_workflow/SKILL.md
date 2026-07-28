---
name: github_actions_workflow
description: Guidelines for creating and updating GitHub Actions workflows, ensuring explicit permissions and pinact SHA locking.
---

# GitHub Actions Workflow Guidelines

## Rules & Best Practices

1. **Explicit Permissions**
   - Always specify `permissions` explicitly at the workflow or job level to enforce the principle of least privilege.

2. **Pin Actions with `pinact`**
   - Use `pinact` to lock GitHub Actions to full commit SHAs and handle version updates automatically.
