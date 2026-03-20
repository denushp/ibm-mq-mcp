package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"ibm-mq-mcp/internal/mq"
	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"
)

var supportedChannelTypes = []string{"SVRCONN", "SDR", "RCVR", "CLNTCONN"}

type Service struct {
	executor mq.Executor
}

func New(executor mq.Executor) *Service {
	return &Service{executor: executor}
}

func (s *Service) GetQueueManager(ctx context.Context, input GetQueueManagerInput) (QueueManagerResult, error) {
	attributes, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_Q_MGR,
		Parameters: []pcf.Parameter{
			pcf.NewIntegerListParameter(mqconst.MQIACF_Q_MGR_ATTRS, []int32{mqconst.MQIACF_ALL}),
		},
	})
	if err != nil {
		return QueueManagerResult{}, err
	}

	status, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_Q_MGR_STATUS,
		Parameters: []pcf.Parameter{
			pcf.NewIntegerListParameter(mqconst.MQIACF_Q_MGR_STATUS_ATTRS, []int32{mqconst.MQIACF_ALL}),
		},
	})
	if err != nil {
		return QueueManagerResult{}, err
	}

	result := QueueManagerResult{
		Summary:    map[string]any{},
		Attributes: map[string]any{},
		Status:     map[string]any{},
	}

	for _, response := range attributes {
		qmgr := qmgrFromValues(response.Parameters)
		result.Summary = mergeMaps(result.Summary, qmgr.Summary)
		result.Attributes = mergeMaps(result.Attributes, qmgr.Attributes)
	}

	for _, response := range status {
		qmgr := qmgrFromValues(response.Parameters)
		result.Summary = mergeMaps(result.Summary, qmgr.Summary)
		result.Status = mergeMaps(result.Status, qmgr.Status)
	}

	return result, nil
}

func (s *Service) ListQueues(ctx context.Context, input ListQueuesInput) (ListQueuesResult, error) {
	definitions, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_Q,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCA_Q_NAME, defaultQueuePattern(input.NamePattern)),
			pcf.NewIntegerParameter(mqconst.MQIA_Q_TYPE, mqconst.MQQT_ALL),
			pcf.NewIntegerListParameter(mqconst.MQIACF_Q_ATTRS, []int32{mqconst.MQIACF_ALL}),
		},
	})
	if err != nil {
		return ListQueuesResult{}, err
	}

	status, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_Q_STATUS,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCA_Q_NAME, defaultQueuePattern(input.NamePattern)),
			pcf.NewIntegerParameter(mqconst.MQIACF_Q_STATUS_TYPE, mqconst.MQIACF_Q_STATUS),
		},
	})
	if err != nil {
		return ListQueuesResult{}, err
	}

	definitionsByName := map[string]QueueItem{}
	for _, response := range definitions {
		queue := queueFromValues(response.Parameters)
		if queueName(queue) == "" {
			continue
		}
		if !input.IncludeSystem && isSystemObject(queueName(queue)) {
			continue
		}
		definitionsByName[queueName(queue)] = queue
	}

	for _, response := range status {
		queue := queueFromValues(response.Parameters)
		name := queueName(queue)
		if name == "" {
			continue
		}
		existing := definitionsByName[name]
		existing.Status = mergeMaps(existing.Status, queue.Status)
		existing.Attributes = mergeMaps(existing.Attributes, queue.Attributes)
		existing.Summary = mergeMaps(existing.Summary, queue.Summary)
		definitionsByName[name] = existing
	}

	items := make([]QueueItem, 0, len(definitionsByName))
	for _, queue := range definitionsByName {
		items = append(items, queue)
	}

	return ListQueuesResult{
		Items: items,
		Count: len(items),
	}, nil
}

func (s *Service) GetQueue(ctx context.Context, input GetQueueInput) (QueueItem, error) {
	result, err := s.ListQueues(ctx, ListQueuesInput{
		Connection:    input.Connection,
		NamePattern:   input.QueueName,
		IncludeSystem: true,
	})
	if err != nil {
		return QueueItem{}, err
	}
	for _, item := range result.Items {
		if queueName(item) == input.QueueName {
			return item, nil
		}
	}
	return QueueItem{}, fmt.Errorf("queue %q not found", input.QueueName)
}

func (s *Service) CreateLocalQueue(ctx context.Context, input CreateLocalQueueInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	command := pcf.Command{
		Command: mqconst.MQCMD_CREATE_Q,
		Parameters: buildQueueMutationParameters(
			input.QueueName,
			input.MaxDepth,
			input.MaxMessageLength,
			input.PutEnabled,
			input.GetEnabled,
			input.Description,
			input.Trigger,
		),
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, command)
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("createLocalQueue", input.QueueName, responses), nil
}

func (s *Service) BrowseMessages(ctx context.Context, input BrowseMessagesInput) (BrowseMessagesResult, error) {
	maxMessages := input.MaxMessages
	if maxMessages <= 0 {
		maxMessages = 10
	}

	result, err := s.executor.BrowseMessages(ctx, input.Connection, mq.BrowseRequest{
		QueueName:   input.QueueName,
		MaxMessages: maxMessages,
	})
	if err != nil {
		return BrowseMessagesResult{}, err
	}

	messages := make([]BrowseMessageResult, 0, len(result.Messages))
	for _, message := range result.Messages {
		messages = append(messages, BrowseMessageResult{
			Payload: DecodePayload(message.Payload),
		})
	}

	return BrowseMessagesResult{
		QueueName: result.QueueName,
		Messages:  messages,
		Count:     len(messages),
	}, nil
}

func (s *Service) UpdateQueue(ctx context.Context, input UpdateQueueInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	command := pcf.Command{
		Command: mqconst.MQCMD_CHANGE_Q,
		Parameters: buildQueueMutationParameters(
			input.QueueName,
			input.MaxDepth,
			input.MaxMessageLength,
			input.PutEnabled,
			input.GetEnabled,
			input.Description,
			input.Trigger,
		),
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, command)
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("updateQueue", input.QueueName, responses), nil
}

func (s *Service) DeleteQueue(ctx context.Context, input DeleteQueueInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_DELETE_Q,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCA_Q_NAME, input.QueueName),
			pcf.NewIntegerParameter(mqconst.MQIA_Q_TYPE, mqconst.MQQT_LOCAL),
		},
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("deleteQueue", input.QueueName, responses), nil
}

func (s *Service) ClearQueue(ctx context.Context, input ClearQueueInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_CLEAR_Q,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCA_Q_NAME, input.QueueName),
		},
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("clearQueue", input.QueueName, responses), nil
}

func (s *Service) ListChannels(ctx context.Context, input ListChannelsInput) (ListChannelsResult, error) {
	definitions, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_CHANNEL,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, defaultChannelPattern(input.NamePattern)),
			pcf.NewIntegerListParameter(mqconst.MQIACF_CHANNEL_ATTRS, []int32{mqconst.MQIACF_ALL}),
		},
	})
	if err != nil {
		return ListChannelsResult{}, err
	}

	status, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_INQUIRE_CHANNEL_STATUS,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, defaultChannelPattern(input.NamePattern)),
			pcf.NewIntegerParameter(mqconst.MQIACH_CHANNEL_INSTANCE_TYPE, mqconst.MQOT_CURRENT_CHANNEL),
		},
	})
	if err != nil {
		return ListChannelsResult{}, err
	}

	itemsByName := map[string]ChannelItem{}
	for _, response := range definitions {
		channel := channelFromValues(response.Parameters)
		if channelName(channel) == "" {
			continue
		}
		itemsByName[channelName(channel)] = channel
	}
	for _, response := range status {
		channel := channelFromValues(response.Parameters)
		name := channelName(channel)
		if name == "" {
			continue
		}
		existing := itemsByName[name]
		existing.Summary = mergeMaps(existing.Summary, channel.Summary)
		existing.Attributes = mergeMaps(existing.Attributes, channel.Attributes)
		existing.Status = mergeMaps(existing.Status, channel.Status)
		itemsByName[name] = existing
	}

	items := make([]ChannelItem, 0, len(itemsByName))
	for _, item := range itemsByName {
		items = append(items, item)
	}

	return ListChannelsResult{
		Items: items,
		Count: len(items),
	}, nil
}

func (s *Service) GetChannel(ctx context.Context, input GetChannelInput) (ChannelItem, error) {
	result, err := s.ListChannels(ctx, ListChannelsInput{
		Connection:  input.Connection,
		NamePattern: input.ChannelName,
	})
	if err != nil {
		return ChannelItem{}, err
	}
	for _, item := range result.Items {
		if channelName(item) == input.ChannelName {
			return item, nil
		}
	}
	return ChannelItem{}, fmt.Errorf("channel %q not found", input.ChannelName)
}

func (s *Service) CreateChannel(ctx context.Context, input CreateChannelInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}

	channelType, err := parseChannelType(input.ChannelType)
	if err != nil {
		return ActionResult{}, err
	}

	parameters := []pcf.Parameter{
		pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, input.ChannelName),
		pcf.NewIntegerParameter(mqconst.MQIACH_CHANNEL_TYPE, channelType),
		pcf.NewIntegerParameter(mqconst.MQIACH_XMIT_PROTOCOL_TYPE, mqconst.MQXPT_TCP),
	}
	parameters = append(parameters, buildChannelAttributeParameters(input)...)

	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command:    mqconst.MQCMD_CREATE_CHANNEL,
		Parameters: parameters,
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("createChannel", input.ChannelName, responses), nil
}

func (s *Service) DeleteChannel(ctx context.Context, input DeleteChannelInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}

	parameters := []pcf.Parameter{
		pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, input.ChannelName),
	}
	if strings.TrimSpace(input.ChannelType) != "" {
		channelType, err := parseChannelType(input.ChannelType)
		if err != nil {
			return ActionResult{}, err
		}
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIACH_CHANNEL_TYPE, channelType))
	}

	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command:    mqconst.MQCMD_DELETE_CHANNEL,
		Parameters: parameters,
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("deleteChannel", input.ChannelName, responses), nil
}

func (s *Service) StartChannel(ctx context.Context, input StartChannelInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_START_CHANNEL,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, input.ChannelName),
		},
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("startChannel", input.ChannelName, responses), nil
}

func (s *Service) StopChannel(ctx context.Context, input StopChannelInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}
	responses, err := s.executor.RunPCF(ctx, input.Connection, pcf.Command{
		Command: mqconst.MQCMD_STOP_CHANNEL,
		Parameters: []pcf.Parameter{
			pcf.NewStringParameter(mqconst.MQCACH_CHANNEL_NAME, input.ChannelName),
		},
	})
	if err != nil {
		return ActionResult{}, err
	}
	return actionResultFromResponses("stopChannel", input.ChannelName, responses), nil
}

func (s *Service) PutTestMessage(ctx context.Context, input PutTestMessageInput) (ActionResult, error) {
	if err := input.Validate(); err != nil {
		return ActionResult{}, err
	}

	payload := []byte(input.PayloadText)
	if strings.TrimSpace(input.PayloadBase64) != "" {
		decoded, err := decodeBase64(input.PayloadBase64)
		if err != nil {
			return ActionResult{}, err
		}
		payload = decoded
	}

	priority := int32(0)
	if input.Priority != nil {
		priority = *input.Priority
	}

	persistence := mqconst.MQPER_NOT_PERSISTENT
	if input.Persistent != nil && *input.Persistent {
		persistence = mqconst.MQPER_PERSISTENT
	}

	result, err := s.executor.PutMessage(ctx, input.Connection, mq.PutRequest{
		QueueName:   input.QueueName,
		Payload:     payload,
		Format:      mqconst.MQFMT_STRING,
		Priority:    priority,
		Persistence: persistence,
	})
	if err != nil {
		return ActionResult{}, err
	}

	return ActionResult{
		Summary: map[string]any{
			"queueName":     result.QueueName,
			"messageId":     result.MessageID,
			"correlationId": result.CorrelationID,
		},
	}, nil
}

func queueFromValues(values []pcf.Value) QueueItem {
	item := QueueItem{
		Summary:    map[string]any{},
		Attributes: map[string]any{},
		Status:     map[string]any{},
	}

	for _, value := range values {
		name := mqconst.ParameterName(value.Parameter)
		if name == "" {
			continue
		}

		switch value.Type {
		case mqconst.MQCFT_STRING:
			item.Attributes[name] = firstString(value.Strings)
			if name == "name" {
				item.Summary["name"] = firstString(value.Strings)
			}
		case mqconst.MQCFT_INTEGER, mqconst.MQCFT_INTEGER_LIST:
			integer := firstInteger(value.Integers)
			switch value.Parameter {
			case mqconst.MQIA_CURRENT_Q_DEPTH, mqconst.MQIA_OPEN_INPUT_COUNT, mqconst.MQIA_OPEN_OUTPUT_COUNT, mqconst.MQIACF_UNCOMMITTED_MSGS:
				item.Status[name] = integer
			case mqconst.MQIA_Q_TYPE:
				item.Attributes[name] = mqconst.QueueTypeName(integer)
				item.Summary["queueType"] = mqconst.QueueTypeName(integer)
			case mqconst.MQIA_INHIBIT_GET:
				item.Attributes[name] = integer == int64(mqconst.MQQA_GET_ALLOWED)
			case mqconst.MQIA_INHIBIT_PUT:
				item.Attributes[name] = integer == int64(mqconst.MQQA_PUT_ALLOWED)
			default:
				item.Attributes[name] = integer
			}
		}
	}

	return item
}

func qmgrFromValues(values []pcf.Value) QueueManagerResult {
	result := QueueManagerResult{
		Summary:    map[string]any{},
		Attributes: map[string]any{},
		Status:     map[string]any{},
	}

	for _, value := range values {
		name := mqconst.ParameterName(value.Parameter)
		if name == "" {
			continue
		}

		switch value.Type {
		case mqconst.MQCFT_STRING:
			result.Attributes[name] = firstString(value.Strings)
			if name == "queueManager" {
				result.Summary["name"] = firstString(value.Strings)
			}
		case mqconst.MQCFT_INTEGER, mqconst.MQCFT_INTEGER_LIST:
			result.Attributes[name] = firstInteger(value.Integers)
		}
	}

	return result
}

func channelFromValues(values []pcf.Value) ChannelItem {
	item := ChannelItem{
		Summary:    map[string]any{},
		Attributes: map[string]any{},
		Status:     map[string]any{},
	}

	for _, value := range values {
		name := mqconst.ParameterName(value.Parameter)
		if name == "" {
			continue
		}

		switch value.Type {
		case mqconst.MQCFT_STRING:
			item.Attributes[name] = firstString(value.Strings)
			if name == "name" {
				item.Summary["name"] = firstString(value.Strings)
			}
		case mqconst.MQCFT_INTEGER, mqconst.MQCFT_INTEGER_LIST:
			integer := firstInteger(value.Integers)
			switch value.Parameter {
			case mqconst.MQIACH_CHANNEL_STATUS, mqconst.MQIACH_MSGS, mqconst.MQIACH_BYTES_SENT, mqconst.MQIACH_BYTES_RCVD:
				item.Status[name] = integer
			case mqconst.MQIACH_CHANNEL_TYPE:
				item.Attributes[name] = mqconst.ChannelTypeName(integer)
				item.Summary["channelType"] = mqconst.ChannelTypeName(integer)
			default:
				item.Attributes[name] = integer
			}
		}
	}

	return item
}

func defaultQueuePattern(pattern string) string {
	if strings.TrimSpace(pattern) == "" {
		return "*"
	}
	return strings.TrimSpace(pattern)
}

func defaultChannelPattern(pattern string) string {
	if strings.TrimSpace(pattern) == "" {
		return "*"
	}
	return strings.TrimSpace(pattern)
}

func queueName(item QueueItem) string {
	if name, ok := item.Summary["name"].(string); ok {
		return name
	}
	if name, ok := item.Attributes["name"].(string); ok {
		return name
	}
	return ""
}

func channelName(item ChannelItem) string {
	if name, ok := item.Summary["name"].(string); ok {
		return name
	}
	if name, ok := item.Attributes["name"].(string); ok {
		return name
	}
	return ""
}

func isSystemObject(name string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(name)), "SYSTEM.")
}

func mergeMaps(left map[string]any, right map[string]any) map[string]any {
	if left == nil && right == nil {
		return nil
	}
	merged := map[string]any{}
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}
	return merged
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func firstInteger(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}

func decodeBase64(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode payloadBase64: %w", err)
	}
	return decoded, nil
}

func buildQueueMutationParameters(
	queueName string,
	maxDepth *int,
	maxMessageLength *int,
	putEnabled *bool,
	getEnabled *bool,
	description *string,
	trigger *TriggerSettings,
) []pcf.Parameter {
	parameters := []pcf.Parameter{
		pcf.NewStringParameter(mqconst.MQCA_Q_NAME, queueName),
		pcf.NewIntegerParameter(mqconst.MQIA_Q_TYPE, mqconst.MQQT_LOCAL),
	}

	if maxDepth != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_MAX_Q_DEPTH, int32(*maxDepth)))
	}
	if maxMessageLength != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_MAX_MSG_LENGTH, int32(*maxMessageLength)))
	}
	if putEnabled != nil {
		value := mqconst.MQQA_PUT_INHIBITED
		if *putEnabled {
			value = mqconst.MQQA_PUT_ALLOWED
		}
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_INHIBIT_PUT, value))
	}
	if getEnabled != nil {
		value := mqconst.MQQA_GET_INHIBITED
		if *getEnabled {
			value = mqconst.MQQA_GET_ALLOWED
		}
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_INHIBIT_GET, value))
	}
	if description != nil && strings.TrimSpace(*description) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCA_Q_DESC, strings.TrimSpace(*description)))
	}
	if trigger != nil {
		parameters = append(parameters, buildTriggerParameters(*trigger)...)
	}
	return parameters
}

func buildTriggerParameters(trigger TriggerSettings) []pcf.Parameter {
	parameters := []pcf.Parameter{}
	if trigger.Enabled != nil {
		value := mqconst.MQTC_OFF
		if *trigger.Enabled {
			value = mqconst.MQTC_ON
		}
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_TRIGGER_CONTROL, value))
	}
	if trigger.Type != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_TRIGGER_TYPE, parseTriggerType(*trigger.Type)))
	}
	if trigger.Depth != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIA_TRIGGER_DEPTH, int32(*trigger.Depth)))
	}
	if trigger.Data != nil && strings.TrimSpace(*trigger.Data) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCA_TRIGGER_DATA, strings.TrimSpace(*trigger.Data)))
	}
	if trigger.ProcessName != nil && strings.TrimSpace(*trigger.ProcessName) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCA_PROCESS_NAME, strings.TrimSpace(*trigger.ProcessName)))
	}
	if trigger.InitiationQueue != nil && strings.TrimSpace(*trigger.InitiationQueue) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCA_INITIATION_Q_NAME, strings.TrimSpace(*trigger.InitiationQueue)))
	}
	return parameters
}

func buildChannelAttributeParameters(input CreateChannelInput) []pcf.Parameter {
	parameters := []pcf.Parameter{}
	if strings.TrimSpace(input.Description) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCACH_DESC, strings.TrimSpace(input.Description)))
	}
	if strings.TrimSpace(input.ConnectionName) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCACH_CONNECTION_NAME, strings.TrimSpace(input.ConnectionName)))
	}
	if strings.TrimSpace(input.XmitQueueName) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCACH_XMIT_Q_NAME, strings.TrimSpace(input.XmitQueueName)))
	}
	if strings.TrimSpace(input.MCAUserID) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCACH_MCA_USER_ID, strings.TrimSpace(input.MCAUserID)))
	}
	if strings.TrimSpace(input.SSLCipherSpec) != "" {
		parameters = append(parameters, pcf.NewStringParameter(mqconst.MQCACH_SSL_CIPHER_SPEC, strings.TrimSpace(input.SSLCipherSpec)))
	}
	if input.BatchSize != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIACH_BATCH_SIZE, int32(*input.BatchSize)))
	}
	if input.DisconnectInterval != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIACH_DISC_INTERVAL, int32(*input.DisconnectInterval)))
	}
	if input.HeartbeatInterval != nil {
		parameters = append(parameters, pcf.NewIntegerParameter(mqconst.MQIACH_HB_INTERVAL, int32(*input.HeartbeatInterval)))
	}
	return parameters
}

func parseTriggerType(triggerType string) int32 {
	switch strings.ToUpper(strings.TrimSpace(triggerType)) {
	case "FIRST":
		return mqconst.MQTT_FIRST
	case "EVERY":
		return mqconst.MQTT_EVERY
	case "DEPTH":
		return mqconst.MQTT_DEPTH
	default:
		return mqconst.MQTT_NONE
	}
}

func parseChannelType(channelType string) (int32, error) {
	channelType = strings.ToUpper(strings.TrimSpace(channelType))
	if !slices.Contains(supportedChannelTypes, channelType) {
		return 0, fmt.Errorf("unsupported channelType %q", channelType)
	}

	switch channelType {
	case "SDR":
		return mqconst.MQCHT_SENDER, nil
	case "RCVR":
		return mqconst.MQCHT_RECEIVER, nil
	case "CLNTCONN":
		return mqconst.MQCHT_CLNTCONN, nil
	case "SVRCONN":
		return mqconst.MQCHT_SVRCONN, nil
	default:
		return 0, fmt.Errorf("unsupported channelType %q", channelType)
	}
}

func actionResultFromResponses(action string, objectName string, responses []pcf.Response) ActionResult {
	result := ActionResult{
		Summary: map[string]any{
			"action": action,
			"name":   objectName,
		},
		RawCodes: map[string]int32{},
	}
	if len(responses) > 0 {
		last := responses[len(responses)-1]
		result.RawCodes["compCode"] = last.Header.CompCode
		result.RawCodes["reason"] = last.Header.Reason
	}
	return result
}

var errUnsupportedExecutor = fmt.Errorf("MQ executor is not configured")
