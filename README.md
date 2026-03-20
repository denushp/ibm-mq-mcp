# IBM MQ MCP Server

A Go-based MCP (Model Context Protocol) server for IBM MQ management via PCF (Programmable Command Format) commands over stdio.

This repository is designed for AI-assisted IBM MQ work:

- run a local MCP server that talks to your queue manager
- install a shared IBM MQ skill for Codex or Claude
- let the agent automatically reach for the local `ibm-mq` MCP server when the conversation turns into queue manager, queue, channel, or message work

## What You Get

- 15 IBM MQ tools exposed through MCP
- a shared skill at `skills/ibm-mq-mcp/` that can be installed into Codex or Claude
- repo guidance files for Claude (`CLAUDE.md`) and Codex-style agents (`AGENTS.md`)
- a project-scoped Claude MCP config in `.mcp.json`

## One-Command Setup

After building `ibm-mq-mcp`, you can install the shared skill plus Codex and Claude MCP registrations with one script:

```bash
./scripts/install-ai-tooling.sh \
  --mq-install-path "$MQ_INSTALL_PATH" \
  --binary "$(pwd)/ibm-mq-mcp"
```

Preview only:

```bash
./scripts/install-ai-tooling.sh \
  --mq-install-path "$MQ_INSTALL_PATH" \
  --binary "$(pwd)/ibm-mq-mcp" \
  --dry-run
```

The installer:

- links the shared skill into `~/.codex/skills/ibm-mq-mcp`
- links the shared skill into `~/.claude/skills/ibm-mq-mcp`
- refreshes the Codex `ibm-mq` MCP entry
- refreshes the Claude `ibm-mq` MCP entry

You can limit installation to one tool:

```bash
./scripts/install-ai-tooling.sh --mq-install-path "$MQ_INSTALL_PATH" --codex-only
./scripts/install-ai-tooling.sh --mq-install-path "$MQ_INSTALL_PATH" --claude-only
```

## 中文快速开始

如果你希望 Claude Code 和 Codex 在本地都能直接调用 IBM MQ MCP，最短路径是：

1. 安装 IBM MQ Client，并设置 `MQ_INSTALL_PATH=/path/to/ibm-mq`
2. 用 `mqclient` tag 编译真实可连接 MQ 的二进制
3. 运行一键安装脚本，把 skill 和 MCP 注册到本机
4. 重启 Codex，重新开一个 Claude Code 会话

示例：

```bash
export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
go build -tags mqclient -o ./ibm-mq-mcp ./cmd/ibm-mq-mcp

./scripts/install-ai-tooling.sh \
  --mq-install-path "$MQ_INSTALL_PATH" \
  --binary "$(pwd)/ibm-mq-mcp"
```

如果你只想在当前仓库里让 Claude 可用，也可以不安装 Claude 的全局 skill，直接依赖仓库里的 `CLAUDE.md` 和 `.mcp.json`。这种情况下至少要保证：

```bash
export MQ_INSTALL_PATH=/path/to/ibm-mq
export IBM_MQ_MCP_BIN="$(pwd)/ibm-mq-mcp"
```

之后就可以直接提类似下面的问题：

- `连接到 QM1 并列出非 SYSTEM 队列`
- `用这个 connection 对象浏览 APP.INPUT 前 5 条消息`
- `检查 TO.QM2 通道状态并总结异常`

## Quick Start

### 1. Build the real IBM MQ-enabled binary

The default build without `mqclient` is useful for tests and CI, but it will not connect to a real IBM MQ queue manager. For actual MQ access, build with the `mqclient` tag.

```bash
git clone https://github.com/denushp/ibm-mq-mcp.git
cd ibm-mq-mcp

export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"

go build -tags mqclient -o ./ibm-mq-mcp ./cmd/ibm-mq-mcp
```

### 2. Put the binary on your PATH

Using a PATH-installed binary keeps the Claude and Codex setup steps portable across machines and shells.

```bash
install -d "$HOME/.local/bin"
install -m 0755 ./ibm-mq-mcp "$HOME/.local/bin/ibm-mq-mcp"
export PATH="$HOME/.local/bin:$PATH"
```

You can verify the binary is reachable with:

```bash
command -v ibm-mq-mcp
```

If you do not want to place the binary on your PATH, you can instead export an explicit override used by the repository's `.mcp.json`:

```bash
export IBM_MQ_MCP_BIN=/full/path/to/ibm-mq-mcp
```

### 3. Install the shared IBM MQ skill

The repository ships one reusable skill directory at `skills/ibm-mq-mcp/`.

For Codex, install it into `~/.codex/skills`:

```bash
mkdir -p "$HOME/.codex/skills"
ln -snf "$(pwd)/skills/ibm-mq-mcp" "$HOME/.codex/skills/ibm-mq-mcp"
```

For Claude Code:

- inside this repository, `CLAUDE.md` already imports the shared skill, so no extra skill install is required
- if you want the same skill available across all Claude projects, also install it into `~/.claude/skills`

```bash
mkdir -p "$HOME/.claude/skills"
ln -snf "$(pwd)/skills/ibm-mq-mcp" "$HOME/.claude/skills/ibm-mq-mcp"
```

After installing the skill:

- restart Codex so it reloads `~/.codex/skills`
- start a new Claude Code session if you installed the personal Claude skill

### 4. Register the MCP server

#### Codex

Register the local stdio server once at user scope:

```bash
codex mcp add ibm-mq \
  --env MQ_INSTALL_PATH="$MQ_INSTALL_PATH" \
  --env DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64" \
  --env LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64" \
  -- ibm-mq-mcp
```

Verify:

```bash
codex mcp get ibm-mq
```

#### Claude Code

This repository already includes a project-scoped `.mcp.json` that uses `ibm-mq-mcp` from your PATH, or `IBM_MQ_MCP_BIN` if you set it, and expands `MQ_INSTALL_PATH` into the runtime library paths. If you launch Claude Code inside this repository, that project MCP config is enough.

If you want IBM MQ available in all Claude projects, add it at user scope:

```bash
claude mcp add ibm-mq --scope user \
  -e MQ_INSTALL_PATH="$MQ_INSTALL_PATH" \
  -e DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64" \
  -e LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64" \
  -- ibm-mq-mcp
```

Verify:

```bash
claude mcp get ibm-mq
```

## How The Auto-Triggering Works

Once the skill is installed and the MCP server is configured:

- Codex can trigger the installed `ibm-mq-mcp` skill when the user asks to connect to IBM MQ, inspect a queue manager, list queues, browse messages, or operate channels
- Claude can use the same skill if you install it under `~/.claude/skills`, and this repository also imports the shared skill from `CLAUDE.md`
- inside this repository, Claude can discover the project-scoped MCP server from `.mcp.json`
- inside this repository, Codex-style agents can pick up repo guidance from `AGENTS.md`

Typical prompts:

- `Connect to QM1 on mq.example.com and list non-system queues`
- `Browse up to 5 messages on APP.INPUT with this connection object`
- `Check channel TO.QM2 and summarize its status`
- `Create a local queue APP.RETRY on QM1`

## Connection Object

Every tool requires a `connection` object:

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
    "keyRepository": "/path/to/keyrepo",
    "certificateLabel": "ibmwebspheremquser",
    "peerName": "CN=mq.example.com"
  }
}
```

Required fields:

- `host`
- `port`
- `channel`
- `queueManager`

Optional fields:

- `user`
- `password`
- `replyModelQueue`
- `tls`

Detailed tool and field notes live in [skills/ibm-mq-mcp/references/connection-and-tools.md](skills/ibm-mq-mcp/references/connection-and-tools.md).

## MCP Tools

Query tools:

- `get_queue_manager`
- `list_queues`
- `get_queue`
- `list_channels`
- `get_channel`

Queue operations:

- `create_local_queue`
- `delete_queue`
- `update_queue`
- `clear_queue`

Channel operations:

- `create_channel`
- `delete_channel`
- `start_channel`
- `stop_channel`

Messaging tools:

- `browse_messages`
- `put_test_message`

## Runtime Notes

- macOS typically needs `DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"`
- Linux typically needs `LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"`
- `browse_messages` is non-destructive and defaults to 10 messages when `maxMessages` is omitted or non-positive
- `put_test_message` accepts exactly one of `payloadText` or `payloadBase64`
- mutation tools return a structured summary plus MQ completion and reason codes rather than a simple boolean

## Development

```bash
# Unit tests with the stub executor
go test ./...

# Tests with real IBM MQ support
export MQ_INSTALL_PATH=/path/to/ibm-mq
export CGO_LDFLAGS="-L$MQ_INSTALL_PATH/lib64"
export DYLD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
export LD_LIBRARY_PATH="$MQ_INSTALL_PATH/lib64"
go test -tags mqclient ./...

# Run with real IBM MQ support
go run -tags mqclient ./cmd/ibm-mq-mcp
```

## License

MIT
