package service

import (
	"context"
	"encoding/base64"
	"testing"

	"ibm-mq-mcp/internal/mq"
)

func TestBrowseMessagesUsesDefaultLimitAndDecodesPayloads(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	executor := &fakeExecutor{
		browseResult: mq.BrowseResult{
			QueueName: "DEV.QUEUE.1",
			Messages: []mq.BrowseMessage{
				{Payload: []byte("hello")},
				{Payload: []byte{0xff, 0x00, 0xfe}},
			},
		},
	}

	svc := New(executor)
	result, err := svc.BrowseMessages(context.Background(), BrowseMessagesInput{
		Connection: connection,
		QueueName:  "DEV.QUEUE.1",
	})
	if err != nil {
		t.Fatalf("BrowseMessages() error = %v", err)
	}

	if got, want := executor.lastBrowseRequest.MaxMessages, 10; got != want {
		t.Fatalf("MaxMessages = %d, want %d", got, want)
	}

	if got, want := result.Messages[0].Payload.Format, "text"; got != want {
		t.Fatalf("first payload format = %q, want %q", got, want)
	}

	if got, want := result.Messages[1].Payload.Base64, base64.StdEncoding.EncodeToString([]byte{0xff, 0x00, 0xfe}); got != want {
		t.Fatalf("second payload base64 = %q, want %q", got, want)
	}
}

func TestPutTestMessageUsesTextPayloadAndPersistenceFlag(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	persistent := true
	priority := int32(5)
	executor := &fakeExecutor{}

	svc := New(executor)
	_, err = svc.PutTestMessage(context.Background(), PutTestMessageInput{
		Connection:  connection,
		QueueName:   "DEV.QUEUE.1",
		PayloadText: "hello mq",
		Persistent:  &persistent,
		Priority:    &priority,
	})
	if err != nil {
		t.Fatalf("PutTestMessage() error = %v", err)
	}

	if got, want := string(executor.lastPutRequest.Payload), "hello mq"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}

	if got, want := executor.lastPutRequest.Priority, priority; got != want {
		t.Fatalf("priority = %d, want %d", got, want)
	}

	if executor.lastPutRequest.Persistence == 0 {
		t.Fatal("persistence should be set when Persistent=true")
	}
}
