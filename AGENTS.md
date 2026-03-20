# AGENTS.md

Project guidance for Codex-style agents and other tools that read `AGENTS.md`.

## Project Overview

This repository provides a local stdio MCP server for IBM MQ management through PCF commands.

When the user asks to work with IBM MQ, prefer the local `ibm-mq` MCP server over ad hoc shell commands.

Typical IBM MQ tasks:

- connect to a queue manager
- inspect queue manager status
- list or inspect queues
- browse messages
- create, update, clear, or delete queues
- list, inspect, create, start, stop, or delete channels
- put a test message on a queue

## Shared Skill

The repository ships a reusable skill at `skills/ibm-mq-mcp/`.

- Codex users can install it by linking that directory into `~/.codex/skills/ibm-mq-mcp`
- Claude users can install the same directory into `~/.claude/skills/ibm-mq-mcp`
- detailed setup instructions live in `README.md`

## Operating Rules

- Ensure the local `ibm-mq` MCP server is configured before attempting IBM MQ work.
- Require a `connection` object with `host`, `port`, `channel`, and `queueManager`.
- Prefer read-only operations first:
  - `get_queue_manager`
  - `list_queues`
  - `get_queue`
  - `list_channels`
  - `get_channel`
  - `browse_messages`
- Confirm destructive intent before `delete_queue`, `clear_queue`, or `delete_channel`.
- Use `browse_messages` for inspection because it does not consume messages.
- Use `put_test_message` for validation or smoke tests.

## Build And Test

```bash
go build ./cmd/ibm-mq-mcp

export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
go build -tags mqclient ./cmd/ibm-mq-mcp

go test ./...

export DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
export LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
go test -tags mqclient ./...
```
