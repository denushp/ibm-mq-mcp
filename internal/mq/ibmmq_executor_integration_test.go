//go:build mqclient

package mq

import (
	"context"
	"os"
	"strconv"
	"testing"

	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"
)

func TestIBMMQExecutorRunPCFIntegration(t *testing.T) {
	if os.Getenv("IBM_MQ_INTEGRATION") == "" {
		t.Skip("set IBM_MQ_INTEGRATION=1 to enable integration tests")
	}

	port, err := strconv.Atoi(os.Getenv("IBM_MQ_PORT"))
	if err != nil {
		t.Fatalf("invalid IBM_MQ_PORT: %v", err)
	}

	executor := NewIBMMQExecutor()
	connection, err := (ConnectionParams{
		Host:         os.Getenv("IBM_MQ_HOST"),
		Port:         port,
		Channel:      os.Getenv("IBM_MQ_CHANNEL"),
		QueueManager: os.Getenv("IBM_MQ_QMGR"),
		User:         os.Getenv("IBM_MQ_USER"),
		Password:     os.Getenv("IBM_MQ_PASSWORD"),
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	responses, err := executor.RunPCF(context.Background(), connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_Q_MGR,
		Parameters: []pcf.Parameter{
			pcf.NewIntegerListParameter(mqconst.MQIACF_Q_MGR_ATTRS, []int32{mqconst.MQIACF_ALL}),
		},
	})
	if err != nil {
		t.Fatalf("RunPCF() error = %v", err)
	}

	if len(responses) == 0 {
		t.Fatal("RunPCF() returned no responses")
	}
}
