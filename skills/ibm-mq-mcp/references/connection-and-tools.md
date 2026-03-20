# IBM MQ Connection And Tools

## Connection Schema

Every tool call requires this `connection` object shape:

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
- `tls.cipherSpec`
- `tls.keyRepository`
- `tls.certificateLabel`
- `tls.peerName`

When `replyModelQueue` is omitted, the server normalizes it to `SYSTEM.DEFAULT.MODEL.QUEUE`.

## Tool Inventory

### Read-oriented tools

- `get_queue_manager`
  - input: `connection`
- `list_queues`
  - input: `connection`, optional `namePattern`, optional `includeSystem`
- `get_queue`
  - input: `connection`, `queueName`
- `list_channels`
  - input: `connection`, optional `namePattern`
- `get_channel`
  - input: `connection`, `channelName`
- `browse_messages`
  - input: `connection`, `queueName`, optional `maxMessages`
  - default: `maxMessages=10` when omitted or non-positive
  - behavior: non-destructive

### Queue mutation tools

- `create_local_queue`
  - input: `connection`, `queueName`
  - optional: `maxDepth`, `maxMessageLength`, `putEnabled`, `getEnabled`, `description`, `trigger`
- `update_queue`
  - input: `connection`, `queueName`
  - optional: `maxDepth`, `maxMessageLength`, `putEnabled`, `getEnabled`, `description`, `trigger`
  - note: at least one supported attribute must be supplied
- `clear_queue`
  - input: `connection`, `queueName`
  - note: destructive
- `delete_queue`
  - input: `connection`, `queueName`
  - note: destructive, local queues only

### Channel tools

- `create_channel`
  - required input: `connection`, `channelName`, `channelType`
  - supported `channelType`: `SVRCONN`, `SDR`, `RCVR`, `CLNTCONN`
  - optional shared fields: `description`, `mcaUserId`, `sslCipherSpec`, `batchSize`, `disconnectInterval`, `heartbeatInterval`
  - `SDR` also requires `connectionName` and `xmitQueueName`
  - `CLNTCONN` also requires `connectionName`
- `delete_channel`
  - input: `connection`, `channelName`
- `start_channel`
  - input: `connection`, `channelName`
- `stop_channel`
  - input: `connection`, `channelName`

### Messaging tools

- `put_test_message`
  - input: `connection`, `queueName`
  - requires exactly one of `payloadText` or `payloadBase64`
  - optional: `priority`, `persistent`

## Queue Trigger Settings

Queue creation and queue update support this nested `trigger` object:

```json
{
  "enabled": true,
  "type": "FIRST",
  "depth": 1,
  "data": "my-trigger-data",
  "processName": "MY.PROCESS",
  "initiationQueue": "SYSTEM.DEFAULT.INITIATION.QUEUE"
}
```

Supported trigger `type` values:

- `FIRST`
- `EVERY`
- `DEPTH`

## Result Shapes

Read-oriented tools return structured JSON describing queue managers, queues, channels, or browsed messages.

Mutation-style tools return an action summary and raw MQ result codes, for example:

```json
{
  "summary": {
    "action": "createLocalQueue",
    "name": "APP.INPUT"
  },
  "rawCodes": {
    "compCode": 0,
    "reason": 0
  }
}
```

Do not assume mutations return a simple `success` boolean.

## Suggested Operating Order

When the request is unclear:

1. Use `get_queue_manager` to confirm the connection works.
2. Use `list_queues` or `list_channels` to narrow down targets.
3. Use `get_queue` or `get_channel` for one-object detail.
4. Use a mutation tool only after the target object and intent are explicit.

For queue inspection:

1. `get_queue`
2. `browse_messages`
3. `update_queue` or `clear_queue` only if clearly requested

For message validation:

1. `put_test_message`
2. `browse_messages` to confirm the payload path if the user wants a safe follow-up inspection
