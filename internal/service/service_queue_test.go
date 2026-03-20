package service

import (
	"context"
	"testing"

	"ibm-mq-mcp/internal/mq"
	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"
)

func TestListQueuesMergesDefinitionAndStatus(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	executor := &fakeExecutor{
		pcfResponses: [][]pcf.Response{
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						stringValue(mqconst.MQCA_Q_NAME, "DEV.QUEUE.1"),
						intValue(mqconst.MQIA_Q_TYPE, int64(mqconst.MQQT_LOCAL)),
						intValue(mqconst.MQIA_MAX_Q_DEPTH, 5000),
					},
				},
			},
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q_STATUS, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						stringValue(mqconst.MQCA_Q_NAME, "DEV.QUEUE.1"),
						intValue(mqconst.MQIA_CURRENT_Q_DEPTH, 27),
						intValue(mqconst.MQIA_OPEN_INPUT_COUNT, 2),
						intValue(mqconst.MQIA_OPEN_OUTPUT_COUNT, 1),
					},
				},
			},
		},
	}

	svc := New(executor)
	result, err := svc.ListQueues(context.Background(), ListQueuesInput{
		Connection: connection,
	})
	if err != nil {
		t.Fatalf("ListQueues() error = %v", err)
	}

	if got, want := result.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}

	queue := result.Items[0]
	if got, want := queue.Summary["name"], "DEV.QUEUE.1"; got != want {
		t.Fatalf("summary name = %#v, want %#v", got, want)
	}

	if got, want := queue.Status["currentDepth"], int64(27); got != want {
		t.Fatalf("currentDepth = %#v, want %#v", got, want)
	}
}

func TestUpdateQueueRejectsUnsupportedAttributes(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	svc := New(&fakeExecutor{})
	_, err = svc.UpdateQueue(context.Background(), UpdateQueueInput{
		Connection: connection,
		QueueName:  "DEV.QUEUE.1",
		Unsupported: map[string]any{
			"foo": "bar",
		},
	})
	if err == nil {
		t.Fatal("UpdateQueue() expected an error for unsupported attributes")
	}
}

type fakeExecutor struct {
	pcfResponses [][]pcf.Response
	browseResult mq.BrowseResult
	putResult    mq.PutResult

	lastBrowseRequest mq.BrowseRequest
	lastPutRequest    mq.PutRequest
}

func (f *fakeExecutor) RunPCF(_ context.Context, _ ConnectionParams, _ pcf.Command) ([]pcf.Response, error) {
	if len(f.pcfResponses) == 0 {
		return nil, nil
	}

	response := f.pcfResponses[0]
	f.pcfResponses = f.pcfResponses[1:]
	return response, nil
}

func (f *fakeExecutor) BrowseMessages(_ context.Context, _ ConnectionParams, request mq.BrowseRequest) (mq.BrowseResult, error) {
	f.lastBrowseRequest = request
	if f.browseResult.QueueName == "" {
		return mq.BrowseResult{QueueName: request.QueueName}, nil
	}
	return f.browseResult, nil
}

func (f *fakeExecutor) PutMessage(_ context.Context, _ ConnectionParams, request mq.PutRequest) (mq.PutResult, error) {
	f.lastPutRequest = request
	if f.putResult.QueueName == "" {
		return mq.PutResult{QueueName: request.QueueName}, nil
	}
	return f.putResult, nil
}

func stringValue(parameter int32, value string) pcf.Value {
	return pcf.Value{
		Type:      mqconst.MQCFT_STRING,
		Parameter: parameter,
		Strings:   []string{value},
	}
}

func intValue(parameter int32, value int64) pcf.Value {
	return pcf.Value{
		Type:      mqconst.MQCFT_INTEGER,
		Parameter: parameter,
		Integers:  []int64{value},
	}
}
