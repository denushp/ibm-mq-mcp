package server

import (
	"context"
	"encoding/json"
	"fmt"

	"ibm-mq-mcp/internal/mq"
	"ibm-mq-mcp/internal/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func New(executor mq.Executor) *mcp.Server {
	svc := service.New(executor)
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ibm-mq-mcp",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "get_queue_manager", Description: "Get queue manager properties and status."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.GetQueueManagerInput) (*mcp.CallToolResult, service.QueueManagerResult, error) {
			if err := normalizeGetQueueManagerInput(&input); err != nil {
				return nil, service.QueueManagerResult{}, err
			}
			result, err := svc.GetQueueManager(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "list_queues", Description: "List queues and merge definition and status data."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.ListQueuesInput) (*mcp.CallToolResult, service.ListQueuesResult, error) {
			if err := normalizeListQueuesInput(&input); err != nil {
				return nil, service.ListQueuesResult{}, err
			}
			result, err := svc.ListQueues(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "get_queue", Description: "Get a single queue with merged definition and status data."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.GetQueueInput) (*mcp.CallToolResult, service.QueueItem, error) {
			if err := normalizeGetQueueInput(&input); err != nil {
				return nil, service.QueueItem{}, err
			}
			result, err := svc.GetQueue(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "list_channels", Description: "List channels and merge definition and status data."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.ListChannelsInput) (*mcp.CallToolResult, service.ListChannelsResult, error) {
			if err := normalizeListChannelsInput(&input); err != nil {
				return nil, service.ListChannelsResult{}, err
			}
			result, err := svc.ListChannels(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "get_channel", Description: "Get a single channel with merged definition and status data."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.GetChannelInput) (*mcp.CallToolResult, service.ChannelItem, error) {
			if err := normalizeGetChannelInput(&input); err != nil {
				return nil, service.ChannelItem{}, err
			}
			result, err := svc.GetChannel(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "create_local_queue", Description: "Create a local queue with a curated set of attributes."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.CreateLocalQueueInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeCreateLocalQueueInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.CreateLocalQueue(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "delete_queue", Description: "Delete a queue by name."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.DeleteQueueInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeDeleteQueueInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.DeleteQueue(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "update_queue", Description: "Update a curated set of queue attributes."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.UpdateQueueInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeUpdateQueueInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.UpdateQueue(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "clear_queue", Description: "Clear messages from a queue using PCF."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.ClearQueueInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeClearQueueInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.ClearQueue(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "create_channel", Description: "Create a channel in one of the supported types: SVRCONN, SDR, RCVR, CLNTCONN."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.CreateChannelInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeCreateChannelInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.CreateChannel(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "delete_channel", Description: "Delete a channel by name."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.DeleteChannelInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeDeleteChannelInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.DeleteChannel(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "start_channel", Description: "Start a channel by name."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.StartChannelInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeStartChannelInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.StartChannel(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "stop_channel", Description: "Stop a channel by name."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.StopChannelInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizeStopChannelInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.StopChannel(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "browse_messages", Description: "Browse queue messages without consuming them."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.BrowseMessagesInput) (*mcp.CallToolResult, service.BrowseMessagesResult, error) {
			if err := normalizeBrowseMessagesInput(&input); err != nil {
				return nil, service.BrowseMessagesResult{}, err
			}
			result, err := svc.BrowseMessages(ctx, input)
			return jsonResult(result), result, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "put_test_message", Description: "Put a test message to a queue using text or base64 payloads."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input service.PutTestMessageInput) (*mcp.CallToolResult, service.ActionResult, error) {
			if err := normalizePutTestMessageInput(&input); err != nil {
				return nil, service.ActionResult{}, err
			}
			result, err := svc.PutTestMessage(ctx, input)
			return jsonResult(result), result, err
		})

	return server
}

func jsonResult(value any) *mcp.CallToolResult {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("failed to render JSON result: %v", err)},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(payload)},
		},
	}
}

func normalizeConnection(connection *service.ConnectionParams) error {
	normalized, err := connection.Normalize()
	if err != nil {
		return err
	}
	*connection = normalized
	return nil
}

func normalizeGetQueueManagerInput(input *service.GetQueueManagerInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeListQueuesInput(input *service.ListQueuesInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeGetQueueInput(input *service.GetQueueInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeListChannelsInput(input *service.ListChannelsInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeGetChannelInput(input *service.GetChannelInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeCreateLocalQueueInput(input *service.CreateLocalQueueInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeDeleteQueueInput(input *service.DeleteQueueInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeUpdateQueueInput(input *service.UpdateQueueInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeClearQueueInput(input *service.ClearQueueInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeCreateChannelInput(input *service.CreateChannelInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeDeleteChannelInput(input *service.DeleteChannelInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeStartChannelInput(input *service.StartChannelInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeStopChannelInput(input *service.StopChannelInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizeBrowseMessagesInput(input *service.BrowseMessagesInput) error {
	return normalizeConnection(&input.Connection)
}
func normalizePutTestMessageInput(input *service.PutTestMessageInput) error {
	return normalizeConnection(&input.Connection)
}
