//go:build mqclient

package mq

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"ibm-mq-mcp/internal/mqconst"
	"ibm-mq-mcp/internal/pcf"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
)

const adminCommandQueue = "SYSTEM.ADMIN.COMMAND.QUEUE"

type ibmMQExecutor struct{}

func NewDefaultExecutor() Executor {
	return &ibmMQExecutor{}
}

func NewIBMMQExecutor() Executor {
	return &ibmMQExecutor{}
}

func (e *ibmMQExecutor) RunPCF(_ context.Context, connection ConnectionParams, command pcf.Command) ([]pcf.Response, error) {
	session, err := connect(connection)
	if err != nil {
		return nil, err
	}
	defer session.close()

	commandQueue, err := session.openNamedQueue(adminCommandQueue, ibmmq.MQOO_OUTPUT)
	if err != nil {
		return nil, mapMQError(err, CategoryPCFCommandFailure)
	}
	defer closeObject(&commandQueue)

	replyQueue, err := session.openReplyQueue(connection.ReplyModelQueue)
	if err != nil {
		return nil, mapMQError(err, CategoryPCFCommandFailure)
	}
	defer closeObject(&replyQueue)

	payload, err := pcf.EncodeRequest(command.Command, command.Parameters)
	if err != nil {
		return nil, err
	}

	putMD := ibmmq.NewMQMD()
	putMD.Format = ibmmq.MQFMT_ADMIN
	putMD.ReplyToQ = strings.TrimSpace(replyQueue.Name)
	putMD.MsgType = ibmmq.MQMT_REQUEST
	putMD.Report = ibmmq.MQRO_PASS_DISCARD_AND_EXPIRY

	putPMO := ibmmq.NewMQPMO()
	putPMO.Options = ibmmq.MQPMO_NO_SYNCPOINT |
		ibmmq.MQPMO_NEW_MSG_ID |
		ibmmq.MQPMO_NEW_CORREL_ID |
		ibmmq.MQPMO_FAIL_IF_QUIESCING

	if err := commandQueue.Put(putMD, putPMO, payload); err != nil {
		return nil, mapMQError(err, CategoryPCFCommandFailure)
	}

	responses := make([]pcf.Response, 0, 4)
	for {
		getMD := ibmmq.NewMQMD()
		getGMO := ibmmq.NewMQGMO()
		getGMO.Options = ibmmq.MQGMO_WAIT | ibmmq.MQGMO_CONVERT | ibmmq.MQGMO_FAIL_IF_QUIESCING
		getGMO.WaitInterval = 3 * 1000

		buffer := make([]byte, 0, 64*1024)
		data, _, err := replyQueue.GetSlice(getMD, getGMO, buffer)
		if err != nil {
			mqerr, ok := err.(*ibmmq.MQReturn)
			if ok && mqerr.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
				break
			}
			return nil, mapMQError(err, CategoryPCFCommandFailure)
		}

		response, err := pcf.ParseResponse(data)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
		if response.Header.Control == mqconst.MQCFC_LAST {
			break
		}
	}

	return responses, nil
}

func (e *ibmMQExecutor) BrowseMessages(_ context.Context, connection ConnectionParams, request BrowseRequest) (BrowseResult, error) {
	session, err := connect(connection)
	if err != nil {
		return BrowseResult{}, err
	}
	defer session.close()

	queue, err := session.openNamedQueue(request.QueueName, ibmmq.MQOO_BROWSE|ibmmq.MQOO_INPUT_SHARED|ibmmq.MQOO_INQUIRE)
	if err != nil {
		return BrowseResult{}, mapMQError(err, CategoryBrowseFailure)
	}
	defer closeObject(&queue)

	result := BrowseResult{QueueName: request.QueueName}
	for index := 0; index < request.MaxMessages; index++ {
		getMD := ibmmq.NewMQMD()
		getGMO := ibmmq.NewMQGMO()
		getGMO.Options = ibmmq.MQGMO_NO_WAIT | ibmmq.MQGMO_CONVERT | ibmmq.MQGMO_FAIL_IF_QUIESCING
		if index == 0 {
			getGMO.Options |= ibmmq.MQGMO_BROWSE_FIRST
		} else {
			getGMO.Options |= ibmmq.MQGMO_BROWSE_NEXT
		}

		buffer := make([]byte, 0, 64*1024)
		data, _, err := queue.GetSlice(getMD, getGMO, buffer)
		if err != nil {
			mqerr, ok := err.(*ibmmq.MQReturn)
			if ok && mqerr.MQRC == ibmmq.MQRC_NO_MSG_AVAILABLE {
				break
			}
			return BrowseResult{}, mapMQError(err, CategoryBrowseFailure)
		}

		result.Messages = append(result.Messages, BrowseMessage{
			MessageID:     hex.EncodeToString(getMD.MsgId),
			CorrelationID: hex.EncodeToString(getMD.CorrelId),
			Format:        strings.TrimSpace(getMD.Format),
			Payload:       append([]byte(nil), data...),
			Priority:      getMD.Priority,
			Persistence:   getMD.Persistence,
		})
	}

	return result, nil
}

func (e *ibmMQExecutor) PutMessage(_ context.Context, connection ConnectionParams, request PutRequest) (PutResult, error) {
	session, err := connect(connection)
	if err != nil {
		return PutResult{}, err
	}
	defer session.close()

	queue, err := session.openNamedQueue(request.QueueName, ibmmq.MQOO_OUTPUT)
	if err != nil {
		return PutResult{}, mapMQError(err, CategoryPutFailure)
	}
	defer closeObject(&queue)

	putMD := ibmmq.NewMQMD()
	if strings.TrimSpace(request.Format) != "" {
		putMD.Format = request.Format
	}
	putMD.Priority = request.Priority
	putMD.Persistence = request.Persistence

	putPMO := ibmmq.NewMQPMO()
	putPMO.Options = ibmmq.MQPMO_NO_SYNCPOINT |
		ibmmq.MQPMO_NEW_MSG_ID |
		ibmmq.MQPMO_NEW_CORREL_ID |
		ibmmq.MQPMO_FAIL_IF_QUIESCING

	if err := queue.Put(putMD, putPMO, request.Payload); err != nil {
		return PutResult{}, mapMQError(err, CategoryPutFailure)
	}

	return PutResult{
		QueueName:     request.QueueName,
		MessageID:     hex.EncodeToString(putMD.MsgId),
		CorrelationID: hex.EncodeToString(putMD.CorrelId),
	}, nil
}

type queueManagerSession struct {
	qmgr ibmmq.MQQueueManager
}

func connect(connection ConnectionParams) (*queueManagerSession, error) {
	normalized, err := connection.Normalize()
	if err != nil {
		return nil, err
	}

	cno := ibmmq.NewMQCNO()
	cd := ibmmq.NewMQCD()
	cd.ChannelName = normalized.Channel
	cd.ConnectionName = fmt.Sprintf("%s(%d)", normalized.Host, normalized.Port)
	cno.ClientConn = cd
	cno.Options = ibmmq.MQCNO_CLIENT_BINDING
	cno.ApplName = "ibm-mq-mcp"

	if normalized.TLS != nil {
		if strings.TrimSpace(normalized.TLS.CipherSpec) != "" {
			cd.SSLCipherSpec = strings.TrimSpace(normalized.TLS.CipherSpec)
		}
		if strings.TrimSpace(normalized.TLS.PeerName) != "" {
			cd.SSLPeerName = strings.TrimSpace(normalized.TLS.PeerName)
		}
		if strings.TrimSpace(normalized.TLS.CertificateLabel) != "" {
			cd.CertificateLabel = strings.TrimSpace(normalized.TLS.CertificateLabel)
		}

		sco := ibmmq.NewMQSCO()
		if strings.TrimSpace(normalized.TLS.KeyRepository) != "" {
			sco.KeyRepository = strings.TrimSpace(normalized.TLS.KeyRepository)
		}
		if strings.TrimSpace(normalized.TLS.CertificateLabel) != "" {
			sco.CertificateLabel = strings.TrimSpace(normalized.TLS.CertificateLabel)
		}
		cno.SSLConfig = sco
	}

	if normalized.User != "" {
		csp := ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = normalized.User
		csp.Password = normalized.Password
		cno.SecurityParms = csp
	}

	qmgr, err := ibmmq.Connx(normalized.QueueManager, cno)
	if err != nil {
		return nil, mapMQError(err, CategoryConnectionFailure)
	}

	return &queueManagerSession{qmgr: qmgr}, nil
}

func (s *queueManagerSession) close() {
	_ = s.qmgr.Disc()
}

func (s *queueManagerSession) openNamedQueue(name string, options int32) (ibmmq.MQObject, error) {
	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = name
	return s.qmgr.Open(od, options)
}

func (s *queueManagerSession) openReplyQueue(modelQueue string) (ibmmq.MQObject, error) {
	od := ibmmq.NewMQOD()
	od.ObjectType = ibmmq.MQOT_Q
	od.ObjectName = modelQueue
	od.DynamicQName = "AMQ.*"
	return s.qmgr.Open(od, ibmmq.MQOO_INPUT_SHARED|ibmmq.MQOO_OUTPUT|ibmmq.MQOO_INQUIRE)
}

func closeObject(object *ibmmq.MQObject) {
	_ = object.Close(0)
}

func mapMQError(err error, defaultCategory string) error {
	if err == nil {
		return nil
	}

	mqerr, ok := err.(*ibmmq.MQReturn)
	if !ok {
		return err
	}

	category := defaultCategory
	switch mqerr.MQRC {
	case ibmmq.MQRC_NOT_AUTHORIZED:
		category = CategoryAuthorization
	case ibmmq.MQRC_UNKNOWN_OBJECT_NAME:
		category = CategoryObjectNotFound
	case ibmmq.MQRC_Q_MGR_NOT_AVAILABLE, ibmmq.MQRC_HOST_NOT_AVAILABLE, ibmmq.MQRC_CHANNEL_NOT_AVAILABLE:
		category = CategoryConnectionFailure
	}

	return &OperationError{
		Category: category,
		CompCode: mqerr.MQCC,
		Reason:   mqerr.MQRC,
		Detail:   mqerr.Error(),
	}
}
