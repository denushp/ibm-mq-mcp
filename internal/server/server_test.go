package server

import (
	"context"
	"strings"
	"testing"

	"ibm-mq-mcp/internal/mq"
	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerRegistersExpectedTools(t *testing.T) {
	t.Parallel()

	server := New(&testExecutor{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	t1, t2 := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}

	expected := []string{
		"get_queue_manager",
		"list_queues",
		"get_queue",
		"list_channels",
		"get_channel",
		"create_local_queue",
		"delete_queue",
		"update_queue",
		"clear_queue",
		"create_channel",
		"delete_channel",
		"start_channel",
		"stop_channel",
		"browse_messages",
		"put_test_message",
	}

	if got, want := len(tools.Tools), len(expected); got != want {
		t.Fatalf("len(tools.Tools) = %d, want %d", got, want)
	}
}

func TestServerCallsListQueuesTool(t *testing.T) {
	t.Parallel()

	executor := &testExecutor{
		pcfResponses: [][]pcf.Response{
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						{Type: mqconst.MQCFT_STRING, Parameter: mqconst.MQCA_Q_NAME, Strings: []string{"DEV.QUEUE.1"}},
						{Type: mqconst.MQCFT_INTEGER, Parameter: mqconst.MQIA_Q_TYPE, Integers: []int64{int64(mqconst.MQQT_LOCAL)}},
					},
				},
			},
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q_STATUS, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						{Type: mqconst.MQCFT_STRING, Parameter: mqconst.MQCA_Q_NAME, Strings: []string{"DEV.QUEUE.1"}},
						{Type: mqconst.MQCFT_INTEGER, Parameter: mqconst.MQIA_CURRENT_Q_DEPTH, Integers: []int64{12}},
					},
				},
			},
		},
	}

	server := New(executor)
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	t1, t2 := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_queues",
		Arguments: map[string]any{
			"connection": map[string]any{
				"host":         "mq.example.com",
				"port":         1414,
				"channel":      "SYSTEM.ADMIN.SVRCONN",
				"queueManager": "QM1",
				"user":         "app",
				"password":     "secret",
			},
		},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("CallTool() returned no content")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result content type = %T, want *mcp.TextContent", result.Content[0])
	}

	if !strings.Contains(text.Text, "DEV.QUEUE.1") {
		t.Fatalf("tool text = %q, want queue name", text.Text)
	}
}

type testExecutor struct {
	pcfResponses [][]pcf.Response
}

func (e *testExecutor) RunPCF(_ context.Context, _ mq.ConnectionParams, _ pcf.Command) ([]pcf.Response, error) {
	if len(e.pcfResponses) == 0 {
		return nil, nil
	}
	response := e.pcfResponses[0]
	e.pcfResponses = e.pcfResponses[1:]
	return response, nil
}

func (e *testExecutor) BrowseMessages(context.Context, mq.ConnectionParams, mq.BrowseRequest) (mq.BrowseResult, error) {
	return mq.BrowseResult{}, nil
}

func (e *testExecutor) PutMessage(context.Context, mq.ConnectionParams, mq.PutRequest) (mq.PutResult, error) {
	return mq.PutResult{}, nil
}
