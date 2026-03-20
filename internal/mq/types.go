package mq

import (
	"context"
	"fmt"
	"strings"

	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"
)

const DefaultReplyModelQueue = mqconst.DefaultReplyModelQueue

type ConnectionParams struct {
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Channel         string     `json:"channel"`
	QueueManager    string     `json:"queueManager"`
	User            string     `json:"user"`
	Password        string     `json:"password"`
	ReplyModelQueue string     `json:"replyModelQueue,omitempty"`
	TLS             *TLSParams `json:"tls,omitempty"`
}

type TLSParams struct {
	CipherSpec       string `json:"cipherSpec,omitempty"`
	KeyRepository    string `json:"keyRepository,omitempty"`
	CertificateLabel string `json:"certificateLabel,omitempty"`
	PeerName         string `json:"peerName,omitempty"`
}

func (c ConnectionParams) Normalize() (ConnectionParams, error) {
	var missing []string
	if strings.TrimSpace(c.Host) == "" {
		missing = append(missing, "host")
	}
	if c.Port <= 0 {
		missing = append(missing, "port")
	}
	if strings.TrimSpace(c.Channel) == "" {
		missing = append(missing, "channel")
	}
	if strings.TrimSpace(c.QueueManager) == "" {
		missing = append(missing, "queueManager")
	}
	if len(missing) > 0 {
		return ConnectionParams{}, fmt.Errorf("missing required connection fields: %s", strings.Join(missing, ", "))
	}

	normalized := c
	normalized.Host = strings.TrimSpace(normalized.Host)
	normalized.Channel = strings.TrimSpace(normalized.Channel)
	normalized.QueueManager = strings.TrimSpace(normalized.QueueManager)
	normalized.User = strings.TrimSpace(normalized.User)
	if strings.TrimSpace(normalized.ReplyModelQueue) == "" {
		normalized.ReplyModelQueue = DefaultReplyModelQueue
	} else {
		normalized.ReplyModelQueue = strings.TrimSpace(normalized.ReplyModelQueue)
	}

	return normalized, nil
}

type BrowseRequest struct {
	QueueName   string
	MaxMessages int
}

type BrowseMessage struct {
	MessageID     string
	CorrelationID string
	Format        string
	Payload       []byte
	Priority      int32
	Persistence   int32
}

type BrowseResult struct {
	QueueName string
	Messages  []BrowseMessage
}

type PutRequest struct {
	QueueName   string
	Payload     []byte
	Format      string
	Priority    int32
	Persistence int32
}

type PutResult struct {
	QueueName     string
	MessageID     string
	CorrelationID string
}

type Executor interface {
	RunPCF(context.Context, ConnectionParams, pcf.Command) ([]pcf.Response, error)
	BrowseMessages(context.Context, ConnectionParams, BrowseRequest) (BrowseResult, error)
	PutMessage(context.Context, ConnectionParams, PutRequest) (PutResult, error)
}
