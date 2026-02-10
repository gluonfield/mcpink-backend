---
name: senior-engineer
description: Senior software engineer second-opinion workflow for architecture, implementation plans, and technical tradeoffs. Use when you need an additional rigorous perspective before or during coding, especially for system design, migrations, reliability risks, refactors, or uncertain implementation direction.
---

# Senior Engineer

Run a parallel senior-engineer consultation and use the result to improve the primary solution without blocking main execution.

## Workflow

1. Collect complete context before invocation.
2. Build one self-contained prompt that includes exact instructions and current context.
3. Run `codex exec` with model `gpt-5.3-codex` in parallel (background task or parallel agent).
4. Continue main task while consultation runs.
5. Merge useful recommendations into the implementation and call out any rejected suggestions with a brief reason.

## Prompt Construction Requirements

Include all relevant details in one message because the consulted agent has no memory:

- Objective and success criteria
- Current architecture and constraints
- Relevant file paths and code snippets
- Environment/runtime/tooling details
- Known errors, risks, and open questions
- Tradeoffs already considered
- Output format you want back (for example: findings, recommended approach, concrete patch plan)

Use concrete facts and exact instructions; avoid placeholder context.

## Invocation Requirements

Always invoke with `gpt-5.3-codex` and high reasoning effort:

```bash
codex exec -m gpt-5.3-codex -c model_reasoning_effort=xhigh "<FULL_PROMPT_WITH_ALL_CONTEXT>" 2>/dev/null
```

## Parallel Execution Rules

Run as a parallel agent team/task and do not block the main execution flow.

- Start the consultation in the background or separate parallel task.
- Keep making forward progress on the primary task.
- Consume the consultation result once available and integrate improvements.
- Do not wait idly for the senior-engineer response unless the main task is blocked.

Example non-blocking shell pattern:

```bash
codex exec -m gpt-5.3-codex -c model_reasoning_effort=xhigh "<FULL_PROMPT_WITH_ALL_CONTEXT>" \
  >/tmp/senior-engineer.out 2>/tmp/senior-engineer.err &
```

## Quality Bar

Treat the consultation as an expert peer review, not final authority.

- Verify recommendations against repository constraints.
- Prefer changes that reduce risk and improve maintainability.
- Keep accepted changes explicit in final reasoning and implementation notes.
