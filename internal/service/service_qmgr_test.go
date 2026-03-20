package service

import (
	"context"
	"testing"

	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"
)

func TestGetQueueManagerMergesAttributesAndStatus(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	executor := &fakeExecutor{
		pcfResponses: [][]pcf.Response{
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q_MGR, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						stringValue(mqconst.MQCA_Q_MGR_NAME, "QM1"),
						stringValue(mqconst.MQCA_DEAD_LETTER_Q_NAME, "SYSTEM.DEAD.LETTER.QUEUE"),
						intValue(mqconst.MQIA_COMMAND_LEVEL, 943),
						intValue(mqconst.MQIA_PLATFORM, 3),
						intValue(mqconst.MQIA_MAX_CHANNELS, 200),
					},
				},
			},
			{
				{
					Header: pcf.Header{Command: mqconst.MQCMD_INQUIRE_Q_MGR_STATUS, CompCode: mqconst.MQCC_OK},
					Parameters: []pcf.Value{
						stringValue(mqconst.MQCA_Q_MGR_NAME, "QM1"),
					},
				},
			},
		},
	}

	svc := New(executor)
	result, err := svc.GetQueueManager(context.Background(), GetQueueManagerInput{
		Connection: connection,
	})
	if err != nil {
		t.Fatalf("GetQueueManager() error = %v", err)
	}

	if got, want := result.Summary["name"], "QM1"; got != want {
		t.Fatalf("summary name = %#v, want %#v", got, want)
	}

	if got, want := result.Attributes["deadLetterQueue"], "SYSTEM.DEAD.LETTER.QUEUE"; got != want {
		t.Fatalf("deadLetterQueue = %#v, want %#v", got, want)
	}

	if got, want := result.Attributes["commandLevel"], int64(943); got != want {
		t.Fatalf("commandLevel = %#v, want %#v", got, want)
	}
}
