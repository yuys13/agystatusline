---
name: github_pr_workflow
description: Standard workflow guidelines for creating Pull Requests, including branch creation, commit cleanup, user permission requests before pushing, and monitoring CI status via gh pr checks --watch.
---

# GitHub PR Workflow Guidelines

When developing features or fixing issues that involve creating a Pull Request (PR), follow this mandatory workflow:

## 1. Branch Creation First

Before making any file modifications or running feature implementation commands, always create and switch to a dedicated topic branch:

```bash
git switch -c feature/<feature-name> # or fix/<issue-name>
```

Never perform work directly on the `main` branch.

## 2. Commit Cleanup

When creating commits or before opening a Pull Request:

- Ensure all relevant code verification steps pass (see `code_verification` skill).
- Always read and strictly follow the [git_commit](../git_commit/SKILL.md) skill guidelines when creating git commits (individual staging, Conventional Commits format, and why-focused messages).
- Clean up and organize commits so they are logical, atomic, and follow Conventional Commits.
- **Do NOT amend or rewrite commits once pushed to remote**, unless explicitly instructed by the user.

## 3. Obtain User Approval Before Pushing / PR Creation

Before running `git push` or `gh pr create`:

- Explicitly ask the user for permission to push the branch and open the Pull Request.
- Wait for explicit user confirmation before executing `git push` or `gh pr create`.

## 4. Monitor CI Status Post-PR Creation

Immediately after creating the PR with `gh pr create`:

- Run `gh pr checks --watch` to monitor and verify the execution of GitHub Actions CI checks.
- If any check fails, fetch and inspect the failed logs using `gh run view <run-id> --log-failed` to diagnose and resolve the issue.
