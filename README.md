# Deploy MCP

> **One MCP to deploy them all.** Infrastructure for the agentic era.

---

## Motivation

AI agents can now write complete applications. Claude Code, Cursor, and Windsurf generate production-ready code in minutes. But when it's time to deploy, agents hit a wall:

**The Fragmentation Problem:**
- Need Railway MCP for hosting
- Need Neon MCP for database
- Need separate tools for secrets, DNS, SSL
- Manual wiring of connection strings
- Human must create accounts on each platform

**The Result:** Agents build in seconds, but deploying takes hours of human intervention.

```
# Today
Agent: "I've created your SaaS app. Here's the code."
Human: *creates Railway account*
Human: *creates Neon account*
Human: *deploys app manually*
Human: *provisions database*
Human: *copies connection string*
Human: *sets environment variables*
Human: *configures domain*
Human: *waits for SSL*
→ 2 hours later: "It's live"

# With Deploy MCP
Agent: deploy(repo="github.com/user/my-saas", database={type:"postgres"})
→ 60 seconds later: "https://my-saas.deploy.app is live"
```

---

## Vision

**"Internet for Agents"** — Infrastructure that agents can provision autonomously.

Deploy MCP is a **platform**, not just a tool:
- Users authenticate to **us**
- We provision infrastructure using **our** provider credentials
- Agents deploy with **one command**
- Users never touch provider dashboards

---

## Core Principles

| Principle | Description |
|-----------|-------------|
| **Repo as Identity** | `github.com/user/app` is the natural project key |
| **One Transaction** | App + database + secrets + domain in a single call |
| **Auto-Deploy Default** | Push to GitHub → automatic deployment |
| **Platform Abstraction** | Users never see underlying providers |
| **Right Tool for Job** | Frontend → edge, Backend → containers |

---

## Authentication

Users authenticate to Deploy MCP. We handle all provider credentials internally.

### Flow

```
1. User visits deploy-mcp.dev
2. "Sign in with GitHub" → OAuth
3. We store: user identity + GitHub token (for private repos)
4. Dashboard shows API key: dk_live_abc123...
5. User adds to MCP config:

{
  "mcpServers": {
    "deploy": {
      "command": "deploy-mcp",
      "env": {
        "DEPLOY_API_KEY": "dk_live_abc123..."
      }
    }
  }
}

6. All MCP calls authenticated via API key
7. We use OUR provider credentials behind the scenes
```

### Why GitHub OAuth?

- **Repo is the project key** — Need GitHub identity anyway
- **Private repo access** — OAuth token lets us clone user's repos
- **Verify ownership** — Confirm user owns repo before deploying
- **Familiar** — Every developer has GitHub

---

## Tech Stack

- **Language**: Go
- **MCP Framework**: mcp-go
- **Database**: Postgres
- **Auth**: GitHub OAuth + JWT
