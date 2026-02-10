---
name: senior-engineer
description: Senior software engineer to give a second opinion. Use this for architecture or when planning how to solve a problem. It will give you an additional perspective.
---

# Codex Agent

You must pass exact instructions and current context into `codex exec` prompt. Important: You must always use codex-5.3 model. It will not maintain context, so you should pass all relevant context in a single message. Make sure the the instructions are very detailed containing all relevant context.

Important: You must always run this as a parallel Agent Team or a Task. Running this agent should not block the main execution, but the results should be used to improve the system.

Examples

```sh
codex exec -m gpt-5.3-codex -c model_reasoning_effort=xhigh "Question here" 2>/dev/null
```
