package pcf

import (
	"reflect"
	"testing"

	"ibm-mq-mcp/internal/mqconst"
)

func TestEncodeRequestIncludesHeaderAndParameters(t *testing.T) {
	t.Parallel()

	payload, err := EncodeRequest(mqconst.MQCMD_INQUIRE_Q, []Parameter{
		NewStringParameter(mqconst.MQCA_Q_NAME, "DEV.*"),
		NewIntegerParameter(mqconst.MQIA_Q_TYPE, mqconst.MQQT_ALL),
		NewIntegerListParameter(mqconst.MQIACF_Q_ATTRS, []int32{mqconst.MQIACF_ALL}),
	})
	if err != nil {
		t.Fatalf("EncodeRequest() error = %v", err)
	}

	response, err := ParseResponse(payload)
	if err != nil {
		t.Fatalf("ParseResponse() error = %v", err)
	}

	if got, want := response.Header.Command, int32(mqconst.MQCMD_INQUIRE_Q); got != want {
		t.Fatalf("Header.Command = %d, want %d", got, want)
	}

	if got, want := len(response.Parameters), 3; got != want {
		t.Fatalf("len(response.Parameters) = %d, want %d", got, want)
	}

	if got, want := response.Parameters[0].Strings, []string{"DEV.*"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first parameter strings = %#v, want %#v", got, want)
	}

	if got, want := response.Parameters[1].Integers, []int64{int64(mqconst.MQQT_ALL)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second parameter integers = %#v, want %#v", got, want)
	}
}

func TestParseResponseHandlesStringListAndIntegers(t *testing.T) {
	t.Parallel()

	payload, err := EncodeResponse(Response{
		Header: Header{
			Type:           mqconst.MQCFT_COMMAND_XR,
			Version:        mqconst.MQCFH_VERSION_3,
			Command:        mqconst.MQCMD_INQUIRE_CHANNEL,
			Control:        mqconst.MQCFC_LAST,
			CompCode:       mqconst.MQCC_OK,
			Reason:         mqconst.MQRC_NONE,
			ParameterCount: 3,
		},
		Parameters: []Value{
			{
				Type:      mqconst.MQCFT_STRING,
				Parameter: mqconst.MQCACH_CHANNEL_NAME,
				Strings:   []string{"SYSTEM.ADMIN.SVRCONN"},
			},
			{
				Type:      mqconst.MQCFT_INTEGER,
				Parameter: mqconst.MQIACH_CHANNEL_TYPE,
				Integers:  []int64{int64(mqconst.MQCHT_SVRCONN)},
			},
			{
				Type:      mqconst.MQCFT_INTEGER,
				Parameter: mqconst.MQIACH_XMIT_PROTOCOL_TYPE,
				Integers:  []int64{int64(mqconst.MQXPT_TCP)},
			},
		},
	})
	if err != nil {
		t.Fatalf("EncodeResponse() error = %v", err)
	}

	got, err := ParseResponse(payload)
	if err != nil {
		t.Fatalf("ParseResponse() error = %v", err)
	}

	if got.Header.ParameterCount != 3 {
		t.Fatalf("Header.ParameterCount = %d, want 3", got.Header.ParameterCount)
	}

	if got.Parameters[0].Parameter != mqconst.MQCACH_CHANNEL_NAME {
		t.Fatalf("first parameter = %d, want %d", got.Parameters[0].Parameter, mqconst.MQCACH_CHANNEL_NAME)
	}

	if got.Parameters[1].Integers[0] != int64(mqconst.MQCHT_SVRCONN) {
		t.Fatalf("channel type = %d, want %d", got.Parameters[1].Integers[0], mqconst.MQCHT_SVRCONN)
	}
}
