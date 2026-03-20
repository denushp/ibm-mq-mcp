# CLAUDE.md

Project guidance for Claude Code when working in this repository.

@skills/ibm-mq-mcp/SKILL.md

## Project Overview

This repository provides a local stdio MCP server for IBM MQ administration through PCF commands.

Use the local `ibm-mq` MCP server whenever the task is about:

- connecting to an IBM MQ queue manager
- listing or inspecting queues
- browsing messages
- creating, updating, clearing, or deleting queues
- listing, creating, starting, stopping, or deleting channels
- putting a test message on a queue

## MCP Expectations

- This repository includes a project-scoped `.mcp.json`.
- That config expects `ibm-mq-mcp` to be on `PATH`, or `IBM_MQ_MCP_BIN` to point at the binary explicitly.
- It also expects `MQ_INSTALL_PATH` to be set in the shell that launches Claude Code.
- If the `ibm-mq` MCP server is unavailable, direct the user to the repository README instead of improvising setup steps.

## Operating Rules

- Ask for or assemble a complete `connection` object before using IBM MQ tools.
- Prefer read-only tools first: `get_queue_manager`, `list_queues`, `get_queue`, `list_channels`, `get_channel`, `browse_messages`.
- Confirm destructive intent before `delete_queue`, `clear_queue`, or `delete_channel`.
- Prefer `browse_messages` for inspection because it is non-destructive.
- Prefer `put_test_message` for smoke tests.

## Build And Test Commands

```bash
# Default build for CI and stubbed tests only
go build ./cmd/ibm-mq-mcp

# Real IBM MQ build
export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
go build -tags mqclient ./cmd/ibm-mq-mcp

# Unit tests with the stub executor
go test ./...

# Tests with real IBM MQ support
export DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
export LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
go test -tags mqclient ./...
```

## Architecture

```text
cmd/ibm-mq-mcp/main.go              Entry point
internal/server/server.go           MCP server setup and tool registration
internal/service/service.go         Business logic and PCF orchestration
internal/service/types.go           Tool input and output shapes
internal/mq/types.go                MQ connection and executor contracts
internal/mq/stub_executor.go        Stub executor used without mqclient
internal/mq/ibmmq_executor_mqclient.go  Real IBM MQ executor behind mqclient
internal/pcf/wire.go                PCF wire encoding and decoding
```
