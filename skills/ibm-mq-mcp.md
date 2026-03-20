# IBM MQ MCP Skill

Use this skill when you need to manage IBM MQ queue managers, queues, channels, or messages.

## Connection

All tools require a `connection` object. Get connection details from project environment or config.

```json
{
  "host": "mq.example.com",
  "port": 1414,
  "channel": "SYSTEM.ADMIN.SVRCONN",
  "queueManager": "QM1",
  "user": "",
  "password": ""
}
```

**Required**: `host`, `port`, `channel`, `queueManager`
**Optional**: `user`, `password` (leave empty if no auth), `replyModelQueue`, `tls`

## Tools

### Query
- `get_queue_manager` - Queue manager properties
- `list_queues` - List queues (`namePattern`, `includeSystem`)
- `get_queue` - Single queue details
- `list_channels` - List channels
- `get_channel` - Single channel status

### Queue Mutation
- `create_local_queue` - Create queue
- `update_queue` - Update attributes
- `clear_queue` - Clear messages
- `delete_queue` - Delete queue

### Channel
- `create_channel` - Create channel (type: SVRCONN, SDR, RCVR, CLNTCONN)
- `delete_channel`, `start_channel`, `stop_channel`

### Messaging
- `browse_messages` - Browse without consuming (`maxMessages`)
- `put_test_message` - Put message (`payloadText`, `format`, `persistent`)

## Notes

- `user`/`password` optional - leave empty if no auth required
- `browse_messages` is read-only
- Mutation tools return `{ "success": true/false, "reason": "" }`
