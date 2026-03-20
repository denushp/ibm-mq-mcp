package main

import (
	"context"
	"log"

	"ibm-mq-mcp/internal/mq"
	"ibm-mq-mcp/internal/server"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	srv := server.New(mq.NewDefaultExecutor())
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("ibm-mq-mcp server failed: %v", err)
	}
}
