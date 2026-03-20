//go:build !mqclient

package mq

import (
	"context"

	"ibm-mq-mcp/internal/pcf"
)

type stubExecutor struct{}

func NewDefaultExecutor() Executor {
	return &stubExecutor{}
}

func (s *stubExecutor) RunPCF(context.Context, ConnectionParams, pcf.Command) ([]pcf.Response, error) {
	return nil, unavailableError("IBM MQ support requires building with -tags mqclient on a machine with the IBM MQ client libraries installed")
}

func (s *stubExecutor) BrowseMessages(context.Context, ConnectionParams, BrowseRequest) (BrowseResult, error) {
	return BrowseResult{}, unavailableError("IBM MQ support requires building with -tags mqclient on a machine with the IBM MQ client libraries installed")
}

func (s *stubExecutor) PutMessage(context.Context, ConnectionParams, PutRequest) (PutResult, error) {
	return PutResult{}, unavailableError("IBM MQ support requires building with -tags mqclient on a machine with the IBM MQ client libraries installed")
}
