---
name: retab
description: Use when building with Retab for document parsing, extraction, splitting, editing, classification, workflow runs, consensus, or review. Guides the AI to install Retab tools and read the official docs before implementing.
---

# Retab

Retab is a document automation platform. Use it to build apps and agents that parse documents, extract structured data, split bundles, edit/fill documents, classify documents, partition repeated sections, and run saved workflows.

Retab supports:

- Six document primitives: `parse`, `extract`, `split`, `edit`, `classify`, and `partition`
- Workflows: saved multi-step pipelines that can combine primitives, custom logic, branches, loops, API calls, and review gates
- Review-based routing: workflow runs can pause at review gates and resume after a human decision
- Consensus: run multiple passes and reconcile results when accuracy matters more than latency
- SDKs and REST APIs for Python, Node, Go, Java, and direct HTTP usage
- MCP access so AI agents can inspect and operate Retab through a tool server

Before implementing, install the relevant Retab tools, then read the docs.

## Install

CLI:

```bash
curl -fsSL https://retab.com/install.sh | sh
```

SDKs:

```bash
# Python SDK
pip install retab

# Node SDK
npm install @retab/node

# Go SDK
go get github.com/retab-dev/retab/clients/go

# Java SDK - Maven
mvn dependency:get -Dartifact=com.retab:retab:0.0.12

# Java SDK - Gradle
# Add implementation("com.retab:retab:0.0.12") to build.gradle.kts
```

MCP:

```bash
# Preferred: install the Retab skill and MCP for supported agents
retab setup

# Claude MCP
claude mcp add --transport http retab https://mcp.retab.com/mcp

# Codex MCP
codex mcp add retab --url https://mcp.retab.com/mcp

# Generic MCP config
cat <<'JSON'
{
  "retab": {
    "type": "streamable-http",
    "url": "https://mcp.retab.com/mcp",
    "headers": { "Authorization": "Bearer YOUR_API_KEY" }
  }
}
JSON
```

`retab setup` installs the Retab skill into the universal `.agents/skills/retab`
directory for mainstream agents such as Amp, Antigravity, Cline, Codex, Cursor,
Deep Agents, Dexto, Firebender, Gemini CLI, GitHub Copilot, Kimi Code CLI,
OpenCode, and Warp. It also configures MCP for agents with known MCP config
formats.

## Docs

Read the official docs before choosing API shapes or workflow patterns:

https://docs.retab.com/

## Usage Guidance

- Important: If possible, prefer the CLI. It is much more powerful than the SDK and the MCP.
- Prefer direct primitives for simple one-step document operations.
- Use workflows to compose primitives with each other.
- Use consensus for high-stakes, ambiguous, or low-quality documents.
