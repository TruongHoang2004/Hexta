---
name: basic-code-review
description: Guides the agent on how to perform a basic code review, focusing on readability, potential bugs, performance, and security.
---

# Basic Code Review Skill

When asked to review a piece of code or a file, always follow these steps:

## 1. Readability & Maintainability
- Are the names of variables, functions, and classes clear and descriptive?
- Is the code too complex or excessively nested?
- Are there comments explaining complex logic?

## 2. Potential Bugs
- Is error handling adequate?
- Are there any infinite loops or incorrect exit conditions?
- Are there potential null/undefined pointer dereferences?

## 3. Performance & Security
- Are database queries being executed inside loops (N+1 problem)?
- Is user input validated or sanitized to prevent injections (SQLi, XSS)?

## Review Output Format
Always present the review results in Markdown, divided into the following sections:
1. **Overview**: A brief summary of the code.
2. **The Good**: Parts of the code that are well written.
3. **Areas for Improvement**: List of bugs or areas to improve, accompanied by suggested code snippets (formatted in code blocks).
