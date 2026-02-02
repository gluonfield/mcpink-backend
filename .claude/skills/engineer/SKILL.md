---
name: software-engineer
description: Ask question about the codebase, to debug, come up with new ideas or potential solutions.
---

# Codex Agent

When using codex you must always use gpt-5.2 model. You must pass exact instructions and current context into `codex exec` prompt. It will not maintain context.

Examples

```sh
codex exec -m gpt-5.2 -c model_reasoning_effort=medium "What could be good implementation for x?" 2>/dev/null
```
