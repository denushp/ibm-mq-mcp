---
name: ibm-mq-mcp
description: Manage IBM MQ through a local MCP server for queue manager inspection, queue and channel administration, and safe message testing. Use when the user asks to connect to IBM MQ, inspect a queue manager, list or inspect queues, create or update queues, inspect or control channels, browse messages, or put a test message with a provided MQ connection.
---

# IBM MQ MCP

Use this skill when IBM MQ work should be handled through the local `ibm-mq` MCP server instead of ad hoc shell commands or guessed MQ syntax.

## Quick Start

1. Confirm the `ibm-mq` MCP server is available.
2. Confirm the user supplied a `connection` object or enough details to build one.
3. Prefer read-only tools first:
   - `get_queue_manager`
   - `list_queues`
   - `get_queue`
   - `list_channels`
   - `get_channel`
   - `browse_messages`
4. Use mutation tools only when the intended target is clear:
   - `create_local_queue`
   - `update_queue`
   - `clear_queue`
   - `delete_queue`
   - `create_channel`
   - `delete_channel`
   - `start_channel`
   - `stop_channel`
   - `put_test_message`

If the MCP server or skill is not installed, direct the user to the repository README instead of inventing a fallback workflow.

## Connection Rules

Every tool requires a `connection` object.

Required fields:

```json
{
  "host": "mq.example.com",
  "port": 1414,
  "channel": "SYSTEM.ADMIN.SVRCONN",
  "queueManager": "QM1"
}
```

Optional fields:

```json
{
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

If authentication fields are absent, treat that as intentional unless the user says MQ auth is required.

## Safe Operating Rules

- Prefer discovery before mutation.
- Confirm destructive intent when the request is ambiguous.
- Treat `delete_queue`, `clear_queue`, and `delete_channel` as destructive.
- Use `browse_messages` for inspection because it does not consume messages.
- Use `put_test_message` for smoke tests rather than altering application traffic.
- Summarize the exact queue manager, queue, or channel you are about to affect before risky changes.

## Tool Selection Guide

- Queue manager health or summary -> `get_queue_manager`
- Queue discovery -> `list_queues`
- Queue details -> `get_queue`
- Channel discovery -> `list_channels`
- Channel details -> `get_channel`
- Create a local queue -> `create_local_queue`
- Change supported queue attributes -> `update_queue`
- Clear queue contents -> `clear_queue`
- Delete a local queue -> `delete_queue`
- Create or configure a supported channel -> `create_channel`
- Remove a channel -> `delete_channel`
- Start or stop a channel -> `start_channel`, `stop_channel`
- Browse queue contents safely -> `browse_messages`
- Put a validation or smoke-test message -> `put_test_message`

## Practical Patterns

For discovery:

1. Call `get_queue_manager`.
2. Narrow targets with `list_queues` or `list_channels`.
3. Inspect one object with `get_queue` or `get_channel`.

For queue troubleshooting:

1. Inspect the queue with `get_queue`.
2. Browse a small number of messages with `browse_messages`.
3. Only clear or update the queue if the user explicitly wants that change.

For channel troubleshooting:

1. Inspect the channel with `get_channel`.
2. If the user asks for operational action, use `start_channel` or `stop_channel`.
3. Recreate the channel only when the type and required attributes are known.

## References

Read [references/connection-and-tools.md](references/connection-and-tools.md) when you need:

- the full connection schema
- detailed tool input notes
- channel-type-specific requirements
- message payload rules
- result-shape expectations for mutation tools
