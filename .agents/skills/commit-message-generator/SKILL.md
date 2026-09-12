---
name: commit-message-generator
description: Guides the agent on how to write Conventional Commits messages based on git diffs.
---

# Commit Message Generator Skill

When asked to write a commit message or when automatically committing code, always adhere to the Conventional Commits specification.

## Standard Structure
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Valid Types:
- `feat`: A new feature.
- `fix`: A bug fix.
- `docs`: Documentation only changes.
- `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc).
- `refactor`: A code change that neither fixes a bug nor adds a feature.
- `perf`: A code change that improves performance.
- `test`: Adding missing tests or correcting existing tests.
- `chore`: Changes to the build process or auxiliary tools and libraries.

## Instructions:
1. Carefully read the changes (git diff).
2. Determine the most appropriate `type`.
3. If the change belongs to a specific module or component, add a `scope` (e.g., `feat(auth): add login form`).
4. Write a concise `description` starting with a lowercase verb in the imperative mood (e.g., `add`, not `added` or `adds`).
5. If the change is complex, add a `body` to explain the motivation (WHY) and the contrast with the previous behavior (WHAT changed).
