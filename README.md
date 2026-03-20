# IBM MQ MCP Server

A Go-based MCP (Model Context Protocol) server for IBM MQ management via PCF (Programmable Command Format) commands over stdio.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Agent (Claude Code / Codex CLI)       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  MCP Client                                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  ibm-mq-mcp (global binary)                          │  │
│  │  skills/ibm-mq-mcp.md (project-level skill)           │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (stdio)
┌─────────────────────────────────────────────────────────────┐
│  IBM MQ Queue Manager (user provides connection in tool)    │
└─────────────────────────────────────────────────────────────┘
```

- **MCP Server** (`ibm-mq-mcp`): 全局安装，AI agent 通过 stdio 调用
- **Skill** (`skills/ibm-mq-mcp.md`): 项目内，提供工具使用指南
- **Connection**: 项目内配置，调用工具时通过 `connection` 对象传入

## Prerequisites

- Go 1.21+
- IBM MQ Client 9.3+ (for real MQ connectivity)
- Claude Code or OpenAI Codex CLI

## Installation

### 1. Build MCP Server (global)

```bash
# Clone and build
git clone https://github.com/houpeng/ibm-mq-mcp.git
cd ibm-mq-mcp

# Build
export MQ_INSTALL_PATH=/path/to/ibm-mq  # Your IBM MQ installation
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
export DYLD_LIBRARY_PATH=$MQ_INSTALL_PATH/lib64  # macOS only
go build -tags mqclient -o ibm-mq-mcp ./cmd/ibm-mq-mcp

# Install to user bin (cross-platform)
mkdir -p ~/bin
mv ibm-mq-mcp ~/bin/

# Add ~/bin to PATH if needed
export PATH="$HOME/bin:$PATH"
```

### 2. Configure AI Agent

**Claude Code** - Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "ibm-mq": {
      "command": "~/bin/ibm-mq-mcp"
    }
  }
}
```

**OpenAI Codex CLI** - Add to `~/.codex/config.toml`:

```toml
[mcp_servers.ibm-mq]
type = "stdio"
command = "~/bin/ibm-mq-mcp"
```

### 3. Project Setup (per-project)

Each project using IBM MQ should have its own configuration. The skill file at `skills/ibm-mq-mcp.md` provides guidance to the AI agent on how to use the tools.

## MCP Tools (15 total)

### Query Tools
| Tool | Description |
|------|-------------|
| `get_queue_manager` | Get queue manager properties and status |
| `list_queues` | List queues with optional pattern filter |
| `get_queue` | Get single queue details |
| `list_channels` | List all channels |
| `get_channel` | Get single channel status |

### Queue Operations
| Tool | Description |
|------|-------------|
| `create_local_queue` | Create a local queue |
| `delete_queue` | Delete a queue |
| `update_queue` | Update queue attributes |
| `clear_queue` | Clear all messages from a queue |

### Channel Operations
| Tool | Description |
|------|-------------|
| `create_channel` | Create a channel (SVRCONN, SDR, RCVR, CLNTCONN) |
| `delete_channel` | Delete a channel |
| `start_channel` | Start a channel |
| `stop_channel` | Stop a channel |

### Messaging
| Tool | Description |
|------|-------------|
| `browse_messages` | Browse queue messages (non-destructive) |
| `put_test_message` | Put a test message to a queue |

## Connection Parameters

All tools require a `connection` object (per-project configuration):

```json
{
  "host": "mq.example.com",
  "port": 1414,
  "channel": "SYSTEM.ADMIN.SVRCONN",
  "queueManager": "QM1",
  "user": "",
  "password": "",
  "replyModelQueue": "SYSTEM.DEFAULT.MODEL.QUEUE",
  "tls": {
    "cipherSpec": "TLS_RSA_WITH_AES_128_CBC_SHA256",
    "keyRepository": "/path/to/keyrepo"
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `host` | Yes | MQ server hostname or IP |
| `port` | Yes | Listener port |
| `channel` | Yes | SVRCONN channel name |
| `queueManager` | Yes | Queue manager name |
| `user` | No | Authentication user (leave empty if no auth) |
| `password` | No | Authentication password (leave empty if no auth) |
| `replyModelQueue` | No | Reply queue for PCF commands |
| `tls` | No | TLS configuration |

## Development

```bash
# Run unit tests (stub executor)
go test ./...

# Run tests with real IBM MQ
export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
export DYLD_LIBRARY_PATH=$MQ_INSTALL_PATH/lib64
go test -tags mqclient ./...

# Run the server
go run ./cmd/ibm-mq-mcp
```

## License

MIT
