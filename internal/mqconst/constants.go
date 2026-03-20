package mqconst

const (
	DefaultReplyModelQueue = "SYSTEM.DEFAULT.MODEL.QUEUE"
)

const (
	MQCFT_COMMAND_XR    int32 = 16
	MQCFH_VERSION_3     int32 = 3
	MQCFC_LAST          int32 = 1
	MQCFT_INTEGER       int32 = 3
	MQCFT_STRING        int32 = 4
	MQCFT_INTEGER_LIST  int32 = 5
	MQCFT_STRING_LIST   int32 = 6
	MQCFT_BYTE_STRING   int32 = 9
	MQCFT_INTEGER64     int32 = 23
	MQCFT_INTEGER64LIST int32 = 25
)

const (
	MQCC_OK               int32 = 0
	MQRC_NONE             int32 = 0
	MQRC_NO_MSG_AVAILABLE int32 = 2033
	MQRC_NOT_AUTHORIZED   int32 = 2035
)

const (
	MQCMD_INQUIRE_Q_MGR          int32 = 2
	MQCMD_CHANGE_Q               int32 = 8
	MQCMD_CLEAR_Q                int32 = 9
	MQCMD_CREATE_Q               int32 = 11
	MQCMD_DELETE_Q               int32 = 12
	MQCMD_INQUIRE_Q              int32 = 13
	MQCMD_CREATE_CHANNEL         int32 = 23
	MQCMD_DELETE_CHANNEL         int32 = 24
	MQCMD_INQUIRE_CHANNEL        int32 = 25
	MQCMD_START_CHANNEL          int32 = 28
	MQCMD_STOP_CHANNEL           int32 = 29
	MQCMD_INQUIRE_Q_STATUS       int32 = 41
	MQCMD_INQUIRE_CHANNEL_STATUS int32 = 42
	MQCMD_INQUIRE_Q_MGR_STATUS   int32 = 161
)

const (
	MQCA_DEAD_LETTER_Q_NAME int32 = 2006
	MQCA_INITIATION_Q_NAME  int32 = 2008
	MQCA_PROCESS_NAME       int32 = 2012
	MQCA_Q_DESC             int32 = 2013
	MQCA_Q_MGR_NAME         int32 = 2015
	MQCA_Q_NAME             int32 = 2016
	MQCA_TRIGGER_DATA       int32 = 2023
)

const (
	MQCACH_CHANNEL_NAME    int32 = 3501
	MQCACH_DESC            int32 = 3502
	MQCACH_XMIT_Q_NAME     int32 = 3505
	MQCACH_CONNECTION_NAME int32 = 3506
	MQCACH_MCA_USER_ID     int32 = 3527
	MQCACH_SSL_CIPHER_SPEC int32 = 3544
	MQCACH_CHANNEL_NAMES   int32 = 3512
)

const (
	MQIACF_Q_MGR_ATTRS        int32 = 1001
	MQIACF_Q_ATTRS            int32 = 1002
	MQIACF_ALL                int32 = 1009
	MQIACF_CHANNEL_ATTRS      int32 = 1015
	MQIACF_Q_STATUS_ATTRS     int32 = 1026
	MQIACF_UNCOMMITTED_MSGS   int32 = 1027
	MQIACF_Q_STATUS_TYPE      int32 = 1103
	MQIACF_Q_STATUS           int32 = 1105
	MQIACF_Q_MGR_STATUS_ATTRS int32 = 1229
)

const (
	MQIACH_XMIT_PROTOCOL_TYPE    int32 = 1501
	MQIACH_BATCH_SIZE            int32 = 1502
	MQIACH_DISC_INTERVAL         int32 = 1503
	MQIACH_CHANNEL_TYPE          int32 = 1511
	MQIACH_CHANNEL_INSTANCE_TYPE int32 = 1523
	MQIACH_CHANNEL_STATUS        int32 = 1527
	MQIACH_MSGS                  int32 = 1534
	MQIACH_BYTES_SENT            int32 = 1535
	MQIACH_BYTES_RCVD            int32 = 1536
	MQIACH_HB_INTERVAL           int32 = 1563
)

const (
	MQIA_CURRENT_Q_DEPTH   int32 = 3
	MQIA_INHIBIT_GET       int32 = 9
	MQIA_INHIBIT_PUT       int32 = 10
	MQIA_USAGE             int32 = 12
	MQIA_MAX_MSG_LENGTH    int32 = 13
	MQIA_MAX_Q_DEPTH       int32 = 15
	MQIA_OPEN_INPUT_COUNT  int32 = 17
	MQIA_OPEN_OUTPUT_COUNT int32 = 18
	MQIA_Q_TYPE            int32 = 20
	MQIA_TRIGGER_CONTROL   int32 = 24
	MQIA_TRIGGER_TYPE      int32 = 28
	MQIA_TRIGGER_DEPTH     int32 = 29
	MQIA_COMMAND_LEVEL     int32 = 31
	MQIA_PLATFORM          int32 = 32
	MQIA_MAX_CHANNELS      int32 = 109
)

const (
	MQQT_LOCAL int32 = 1
	MQQT_ALL   int32 = 1001
)

const (
	MQQA_GET_ALLOWED   int32 = 0
	MQQA_GET_INHIBITED int32 = 1
	MQQA_PUT_ALLOWED   int32 = 0
	MQQA_PUT_INHIBITED int32 = 1
)

const (
	MQTC_OFF int32 = 0
	MQTC_ON  int32 = 1
)

const (
	MQTT_NONE  int32 = 0
	MQTT_FIRST int32 = 1
	MQTT_EVERY int32 = 2
	MQTT_DEPTH int32 = 3
)

const (
	MQCHT_SENDER   int32 = 1
	MQCHT_RECEIVER int32 = 3
	MQCHT_CLNTCONN int32 = 6
	MQCHT_SVRCONN  int32 = 7
)

const (
	MQCHS_INACTIVE int32 = 0
	MQCHS_RUNNING  int32 = 3
	MQCHS_STOPPED  int32 = 6
)

const (
	MQOT_CURRENT_CHANNEL int32 = 1011
)

const (
	MQXPT_TCP int32 = 2
)

const (
	MQOO_INPUT_SHARED      int32 = 2
	MQOO_BROWSE            int32 = 8
	MQOO_OUTPUT            int32 = 16
	MQOO_INQUIRE           int32 = 32
	MQOO_SET               int32 = 64
	MQOO_FAIL_IF_QUIESCING int32 = 8192
)

const (
	MQGMO_BROWSE_FIRST      int32 = 16
	MQGMO_BROWSE_NEXT       int32 = 32
	MQGMO_NO_WAIT           int32 = 0
	MQGMO_CONVERT           int32 = 16384
	MQGMO_FAIL_IF_QUIESCING int32 = 8192
)

const (
	MQPMO_FAIL_IF_QUIESCING int32 = 8192
	MQPMO_NO_SYNCPOINT      int32 = 4
	MQPMO_NEW_MSG_ID        int32 = 64
	MQPMO_NEW_CORREL_ID     int32 = 128
)

const (
	MQOT_Q     int32 = 1
	MQOT_Q_MGR int32 = 5
)

const (
	MQCNO_CLIENT_BINDING       int32 = 2048
	MQCSP_AUTH_USER_ID_AND_PWD int32 = 1
)

const (
	MQPER_NOT_PERSISTENT int32 = 0
	MQPER_PERSISTENT     int32 = 1
)

const (
	MQMT_REQUEST                 int32  = 1
	MQRO_PASS_DISCARD_AND_EXPIRY int32  = 16384
	MQFMT_NONE                   string = ""
	MQFMT_ADMIN                  string = "MQADMIN"
	MQFMT_STRING                 string = "MQSTR"
)

var parameterNames = map[int32]string{
	MQCA_DEAD_LETTER_Q_NAME:      "deadLetterQueue",
	MQCA_INITIATION_Q_NAME:       "initiationQueue",
	MQCA_PROCESS_NAME:            "processName",
	MQCA_Q_DESC:                  "description",
	MQCA_Q_MGR_NAME:              "queueManager",
	MQCA_Q_NAME:                  "name",
	MQCA_TRIGGER_DATA:            "triggerData",
	MQCACH_CHANNEL_NAME:          "name",
	MQCACH_DESC:                  "description",
	MQCACH_XMIT_Q_NAME:           "xmitQueueName",
	MQCACH_CONNECTION_NAME:       "connectionName",
	MQCACH_MCA_USER_ID:           "mcaUserId",
	MQCACH_SSL_CIPHER_SPEC:       "sslCipherSpec",
	MQIACH_XMIT_PROTOCOL_TYPE:    "transportType",
	MQIACH_BATCH_SIZE:            "batchSize",
	MQIACH_DISC_INTERVAL:         "disconnectInterval",
	MQIACH_CHANNEL_TYPE:          "channelType",
	MQIACH_CHANNEL_INSTANCE_TYPE: "channelInstanceType",
	MQIACH_CHANNEL_STATUS:        "status",
	MQIACH_MSGS:                  "messages",
	MQIACH_BYTES_SENT:            "bytesSent",
	MQIACH_BYTES_RCVD:            "bytesReceived",
	MQIACH_HB_INTERVAL:           "heartbeatInterval",
	MQIA_CURRENT_Q_DEPTH:         "currentDepth",
	MQIA_INHIBIT_GET:             "getEnabled",
	MQIA_INHIBIT_PUT:             "putEnabled",
	MQIA_USAGE:                   "usage",
	MQIA_MAX_MSG_LENGTH:          "maxMessageLength",
	MQIA_MAX_Q_DEPTH:             "maxDepth",
	MQIA_OPEN_INPUT_COUNT:        "openInputCount",
	MQIA_OPEN_OUTPUT_COUNT:       "openOutputCount",
	MQIA_Q_TYPE:                  "queueType",
	MQIA_TRIGGER_CONTROL:         "triggerControl",
	MQIA_TRIGGER_TYPE:            "triggerType",
	MQIA_TRIGGER_DEPTH:           "triggerDepth",
	MQIA_COMMAND_LEVEL:           "commandLevel",
	MQIA_PLATFORM:                "platform",
	MQIA_MAX_CHANNELS:            "maxChannels",
	MQIACF_UNCOMMITTED_MSGS:      "uncommittedMessages",
}

func ParameterName(parameter int32) string {
	if name, ok := parameterNames[parameter]; ok {
		return name
	}
	return ""
}

func ChannelTypeName(channelType int64) string {
	switch channelType {
	case int64(MQCHT_SENDER):
		return "SDR"
	case int64(MQCHT_RECEIVER):
		return "RCVR"
	case int64(MQCHT_CLNTCONN):
		return "CLNTCONN"
	case int64(MQCHT_SVRCONN):
		return "SVRCONN"
	default:
		return ""
	}
}

func QueueTypeName(queueType int64) string {
	switch queueType {
	case int64(MQQT_LOCAL):
		return "LOCAL"
	case int64(MQQT_ALL):
		return "ALL"
	default:
		return ""
	}
}
