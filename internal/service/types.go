package service

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"ibm-mq-mcp/internal/mq"
)

const DefaultReplyModelQueue = mq.DefaultReplyModelQueue

type ConnectionParams = mq.ConnectionParams
type TLSParams = mq.TLSParams

type PayloadPreview struct {
	Format string `json:"format"`
	Text   string `json:"text,omitempty"`
	Base64 string `json:"base64,omitempty"`
}

type ListQueuesInput struct {
	Connection    ConnectionParams `json:"connection"`
	NamePattern   string           `json:"namePattern,omitempty"`
	IncludeSystem bool             `json:"includeSystem,omitempty"`
}

type GetQueueInput struct {
	Connection ConnectionParams `json:"connection"`
	QueueName  string           `json:"queueName"`
}

type GetQueueManagerInput struct {
	Connection ConnectionParams `json:"connection"`
}

type QueueManagerResult struct {
	Summary    map[string]any `json:"summary"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Status     map[string]any `json:"status,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

type QueueItem struct {
	Summary    map[string]any `json:"summary"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Status     map[string]any `json:"status,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

type ListQueuesResult struct {
	Items           []QueueItem `json:"items"`
	Count           int         `json:"count"`
	PartialFailures []string    `json:"partialFailures,omitempty"`
}

type CreateLocalQueueInput struct {
	Connection       ConnectionParams `json:"connection"`
	QueueName        string           `json:"queueName"`
	MaxDepth         *int             `json:"maxDepth,omitempty"`
	MaxMessageLength *int             `json:"maxMessageLength,omitempty"`
	PutEnabled       *bool            `json:"putEnabled,omitempty"`
	GetEnabled       *bool            `json:"getEnabled,omitempty"`
	Description      *string          `json:"description,omitempty"`
	Trigger          *TriggerSettings `json:"trigger,omitempty"`
}

type UpdateQueueInput struct {
	Connection       ConnectionParams `json:"connection"`
	QueueName        string           `json:"queueName"`
	MaxDepth         *int             `json:"maxDepth,omitempty"`
	MaxMessageLength *int             `json:"maxMessageLength,omitempty"`
	PutEnabled       *bool            `json:"putEnabled,omitempty"`
	GetEnabled       *bool            `json:"getEnabled,omitempty"`
	Description      *string          `json:"description,omitempty"`
	Trigger          *TriggerSettings `json:"trigger,omitempty"`
	Unsupported      map[string]any   `json:"-"`
}

type DeleteQueueInput struct {
	Connection ConnectionParams `json:"connection"`
	QueueName  string           `json:"queueName"`
}

type ClearQueueInput struct {
	Connection ConnectionParams `json:"connection"`
	QueueName  string           `json:"queueName"`
}

type TriggerSettings struct {
	Enabled         *bool   `json:"enabled,omitempty"`
	Type            *string `json:"type,omitempty"`
	Depth           *int    `json:"depth,omitempty"`
	Data            *string `json:"data,omitempty"`
	ProcessName     *string `json:"processName,omitempty"`
	InitiationQueue *string `json:"initiationQueue,omitempty"`
}

type ActionResult struct {
	Summary  map[string]any   `json:"summary"`
	Warnings []string         `json:"warnings,omitempty"`
	RawCodes map[string]int32 `json:"rawCodes,omitempty"`
}

type BrowseMessagesInput struct {
	Connection  ConnectionParams `json:"connection"`
	QueueName   string           `json:"queueName"`
	MaxMessages int              `json:"maxMessages,omitempty"`
}

type BrowseMessageResult struct {
	Payload PayloadPreview `json:"payload"`
}

type BrowseMessagesResult struct {
	QueueName string                `json:"queueName"`
	Messages  []BrowseMessageResult `json:"messages"`
	Count     int                   `json:"count"`
}

type CreateChannelInput struct {
	Connection         ConnectionParams `json:"connection"`
	ChannelName        string           `json:"channelName"`
	ChannelType        string           `json:"channelType"`
	ConnectionName     string           `json:"connectionName,omitempty"`
	XmitQueueName      string           `json:"xmitQueueName,omitempty"`
	Description        string           `json:"description,omitempty"`
	MCAUserID          string           `json:"mcaUserId,omitempty"`
	SSLCipherSpec      string           `json:"sslCipherSpec,omitempty"`
	BatchSize          *int             `json:"batchSize,omitempty"`
	DisconnectInterval *int             `json:"disconnectInterval,omitempty"`
	HeartbeatInterval  *int             `json:"heartbeatInterval,omitempty"`
}

type ListChannelsInput struct {
	Connection  ConnectionParams `json:"connection"`
	NamePattern string           `json:"namePattern,omitempty"`
}

type GetChannelInput struct {
	Connection  ConnectionParams `json:"connection"`
	ChannelName string           `json:"channelName"`
}

type DeleteChannelInput struct {
	Connection  ConnectionParams `json:"connection"`
	ChannelName string           `json:"channelName"`
	ChannelType string           `json:"channelType,omitempty"`
}

type StartChannelInput struct {
	Connection  ConnectionParams `json:"connection"`
	ChannelName string           `json:"channelName"`
}

type StopChannelInput struct {
	Connection  ConnectionParams `json:"connection"`
	ChannelName string           `json:"channelName"`
}

type ChannelItem struct {
	Summary    map[string]any `json:"summary"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Status     map[string]any `json:"status,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

type ListChannelsResult struct {
	Items           []ChannelItem `json:"items"`
	Count           int           `json:"count"`
	PartialFailures []string      `json:"partialFailures,omitempty"`
}

type PutTestMessageInput struct {
	Connection    ConnectionParams `json:"connection"`
	QueueName     string           `json:"queueName"`
	PayloadText   string           `json:"payloadText,omitempty"`
	PayloadBase64 string           `json:"payloadBase64,omitempty"`
	Priority      *int32           `json:"priority,omitempty"`
	Persistent    *bool            `json:"persistent,omitempty"`
}

func (i PutTestMessageInput) Validate() error {
	if strings.TrimSpace(i.QueueName) == "" {
		return fmt.Errorf("queueName is required")
	}
	textSet := strings.TrimSpace(i.PayloadText) != ""
	base64Set := strings.TrimSpace(i.PayloadBase64) != ""
	if textSet == base64Set {
		return fmt.Errorf("exactly one of payloadText or payloadBase64 must be set")
	}
	if base64Set {
		if _, err := base64.StdEncoding.DecodeString(i.PayloadBase64); err != nil {
			return fmt.Errorf("payloadBase64 must be valid base64: %w", err)
		}
	}
	return nil
}

func (i CreateLocalQueueInput) Validate() error {
	if strings.TrimSpace(i.QueueName) == "" {
		return fmt.Errorf("queueName is required")
	}
	return nil
}

func (i UpdateQueueInput) Validate() error {
	if strings.TrimSpace(i.QueueName) == "" {
		return fmt.Errorf("queueName is required")
	}
	if len(i.Unsupported) > 0 {
		return fmt.Errorf("unsupported queue attributes requested: %v", mapsKeys(i.Unsupported))
	}
	if i.MaxDepth == nil && i.MaxMessageLength == nil && i.PutEnabled == nil && i.GetEnabled == nil && i.Description == nil && i.Trigger == nil {
		return fmt.Errorf("at least one supported queue attribute must be provided")
	}
	return nil
}

func (i DeleteQueueInput) Validate() error {
	if strings.TrimSpace(i.QueueName) == "" {
		return fmt.Errorf("queueName is required")
	}
	return nil
}

func (i ClearQueueInput) Validate() error {
	if strings.TrimSpace(i.QueueName) == "" {
		return fmt.Errorf("queueName is required")
	}
	return nil
}

func (i CreateChannelInput) Validate() error {
	if strings.TrimSpace(i.ChannelName) == "" {
		return fmt.Errorf("channelName is required")
	}
	switch strings.ToUpper(strings.TrimSpace(i.ChannelType)) {
	case "SVRCONN", "RCVR":
		return nil
	case "SDR":
		if strings.TrimSpace(i.ConnectionName) == "" {
			return fmt.Errorf("connectionName is required for SDR channels")
		}
		if strings.TrimSpace(i.XmitQueueName) == "" {
			return fmt.Errorf("xmitQueueName is required for SDR channels")
		}
		return nil
	case "CLNTCONN":
		if strings.TrimSpace(i.ConnectionName) == "" {
			return fmt.Errorf("connectionName is required for CLNTCONN channels")
		}
		return nil
	default:
		return fmt.Errorf("unsupported channelType %q", i.ChannelType)
	}
}

func (i DeleteChannelInput) Validate() error {
	if strings.TrimSpace(i.ChannelName) == "" {
		return fmt.Errorf("channelName is required")
	}
	return nil
}

func (i StartChannelInput) Validate() error {
	if strings.TrimSpace(i.ChannelName) == "" {
		return fmt.Errorf("channelName is required")
	}
	return nil
}

func (i StopChannelInput) Validate() error {
	if strings.TrimSpace(i.ChannelName) == "" {
		return fmt.Errorf("channelName is required")
	}
	return nil
}

func DecodePayload(payload []byte) PayloadPreview {
	if utf8.Valid(payload) {
		return PayloadPreview{
			Format: "text",
			Text:   string(payload),
		}
	}
	return PayloadPreview{
		Format: "base64",
		Base64: base64.StdEncoding.EncodeToString(payload),
	}
}

func mapsKeys(input map[string]any) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}
