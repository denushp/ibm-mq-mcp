package pcf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"ibm-mq-mcp/internal/mqconst"
)

const (
	cfhLength             = 36
	integerLength         = 16
	integerListBaseLength = 16
	stringBaseLength      = 20
	stringListBaseLength  = 20
)

var nativeEndian binary.ByteOrder

func init() {
	var i uint16 = 0x1
	buf := *(*[2]byte)(unsafe.Pointer(&i))
	if buf[0] == 0x1 {
		nativeEndian = binary.LittleEndian
		return
	}
	nativeEndian = binary.BigEndian
}

type Header struct {
	Type           int32
	StrucLength    int32
	Version        int32
	Command        int32
	MsgSeqNumber   int32
	Control        int32
	CompCode       int32
	Reason         int32
	ParameterCount int32
}

type Value struct {
	Type      int32
	Parameter int32
	Strings   []string
	Integers  []int64
}

type Response struct {
	Header     Header
	Parameters []Value
}

type Command struct {
	Command    int32
	Parameters []Parameter
}

type Parameter interface {
	encode() ([]byte, error)
}

type stringParameter struct {
	parameter int32
	value     string
}

type integerParameter struct {
	parameter int32
	value     int32
}

type integerListParameter struct {
	parameter int32
	values    []int32
}

func NewStringParameter(parameter int32, value string) Parameter {
	return stringParameter{parameter: parameter, value: value}
}

func NewIntegerParameter(parameter int32, value int32) Parameter {
	return integerParameter{parameter: parameter, value: value}
}

func NewIntegerListParameter(parameter int32, values []int32) Parameter {
	copied := make([]int32, len(values))
	copy(copied, values)
	return integerListParameter{parameter: parameter, values: copied}
}

func EncodeRequest(command int32, parameters []Parameter) ([]byte, error) {
	body := make([]byte, 0, 128)
	for _, parameter := range parameters {
		chunk, err := parameter.encode()
		if err != nil {
			return nil, err
		}
		body = append(body, chunk...)
	}

	header := Header{
		Type:           mqconst.MQCFT_COMMAND_XR,
		StrucLength:    cfhLength,
		Version:        mqconst.MQCFH_VERSION_3,
		Command:        command,
		MsgSeqNumber:   1,
		Control:        mqconst.MQCFC_LAST,
		CompCode:       mqconst.MQCC_OK,
		Reason:         mqconst.MQRC_NONE,
		ParameterCount: int32(len(parameters)),
	}

	return append(encodeHeader(header), body...), nil
}

func EncodeResponse(response Response) ([]byte, error) {
	if response.Header.StrucLength == 0 {
		response.Header.StrucLength = cfhLength
	}
	if response.Header.Type == 0 {
		response.Header.Type = mqconst.MQCFT_COMMAND_XR
	}
	if response.Header.Version == 0 {
		response.Header.Version = mqconst.MQCFH_VERSION_3
	}
	if response.Header.MsgSeqNumber == 0 {
		response.Header.MsgSeqNumber = 1
	}
	if response.Header.Control == 0 {
		response.Header.Control = mqconst.MQCFC_LAST
	}
	if response.Header.ParameterCount == 0 {
		response.Header.ParameterCount = int32(len(response.Parameters))
	}

	body := make([]byte, 0, 128)
	for _, parameter := range response.Parameters {
		chunk, err := encodeValue(parameter)
		if err != nil {
			return nil, err
		}
		body = append(body, chunk...)
	}

	return append(encodeHeader(response.Header), body...), nil
}

func ParseResponse(payload []byte) (Response, error) {
	if len(payload) < cfhLength {
		return Response{}, errors.New("payload too short for MQCFH header")
	}

	header := parseHeader(payload[:cfhLength])
	offset := cfhLength
	parameters := make([]Value, 0, header.ParameterCount)
	for offset < len(payload) {
		value, n, err := parseValue(payload[offset:])
		if err != nil {
			return Response{}, err
		}
		parameters = append(parameters, value)
		offset += n
	}

	return Response{
		Header:     header,
		Parameters: parameters,
	}, nil
}

func encodeHeader(header Header) []byte {
	buf := make([]byte, cfhLength)
	nativeEndian.PutUint32(buf[0:], uint32(header.Type))
	nativeEndian.PutUint32(buf[4:], uint32(header.StrucLength))
	nativeEndian.PutUint32(buf[8:], uint32(header.Version))
	nativeEndian.PutUint32(buf[12:], uint32(header.Command))
	nativeEndian.PutUint32(buf[16:], uint32(header.MsgSeqNumber))
	nativeEndian.PutUint32(buf[20:], uint32(header.Control))
	nativeEndian.PutUint32(buf[24:], uint32(header.CompCode))
	nativeEndian.PutUint32(buf[28:], uint32(header.Reason))
	nativeEndian.PutUint32(buf[32:], uint32(header.ParameterCount))
	return buf
}

func parseHeader(payload []byte) Header {
	return Header{
		Type:           int32(nativeEndian.Uint32(payload[0:])),
		StrucLength:    int32(nativeEndian.Uint32(payload[4:])),
		Version:        int32(nativeEndian.Uint32(payload[8:])),
		Command:        int32(nativeEndian.Uint32(payload[12:])),
		MsgSeqNumber:   int32(nativeEndian.Uint32(payload[16:])),
		Control:        int32(nativeEndian.Uint32(payload[20:])),
		CompCode:       int32(nativeEndian.Uint32(payload[24:])),
		Reason:         int32(nativeEndian.Uint32(payload[28:])),
		ParameterCount: int32(nativeEndian.Uint32(payload[32:])),
	}
}

func encodeValue(value Value) ([]byte, error) {
	switch value.Type {
	case mqconst.MQCFT_INTEGER:
		if len(value.Integers) != 1 {
			return nil, errors.New("integer PCF values require exactly one integer")
		}
		buf := make([]byte, integerLength)
		nativeEndian.PutUint32(buf[0:], uint32(mqconst.MQCFT_INTEGER))
		nativeEndian.PutUint32(buf[4:], uint32(integerLength))
		nativeEndian.PutUint32(buf[8:], uint32(value.Parameter))
		nativeEndian.PutUint32(buf[12:], uint32(int32(value.Integers[0])))
		return buf, nil
	case mqconst.MQCFT_INTEGER_LIST:
		buf := make([]byte, integerListBaseLength+(len(value.Integers)*4))
		nativeEndian.PutUint32(buf[0:], uint32(mqconst.MQCFT_INTEGER_LIST))
		nativeEndian.PutUint32(buf[4:], uint32(len(buf)))
		nativeEndian.PutUint32(buf[8:], uint32(value.Parameter))
		nativeEndian.PutUint32(buf[12:], uint32(len(value.Integers)))
		offset := integerListBaseLength
		for _, integer := range value.Integers {
			nativeEndian.PutUint32(buf[offset:], uint32(int32(integer)))
			offset += 4
		}
		return buf, nil
	case mqconst.MQCFT_STRING:
		if len(value.Strings) != 1 {
			return nil, errors.New("string PCF values require exactly one string")
		}
		length := roundTo4(len(value.Strings[0]))
		buf := make([]byte, stringBaseLength+length)
		nativeEndian.PutUint32(buf[0:], uint32(mqconst.MQCFT_STRING))
		nativeEndian.PutUint32(buf[4:], uint32(len(buf)))
		nativeEndian.PutUint32(buf[8:], uint32(value.Parameter))
		nativeEndian.PutUint32(buf[12:], uint32(1208))
		nativeEndian.PutUint32(buf[16:], uint32(len(value.Strings[0])))
		copy(buf[stringBaseLength:], []byte(value.Strings[0]))
		return buf, nil
	case mqconst.MQCFT_STRING_LIST:
		if len(value.Strings) == 0 {
			return nil, errors.New("string list PCF values require at least one string")
		}
		stringLength := roundTo4(len(value.Strings[0]))
		buf := make([]byte, stringListBaseLength+(len(value.Strings)*stringLength))
		nativeEndian.PutUint32(buf[0:], uint32(mqconst.MQCFT_STRING_LIST))
		nativeEndian.PutUint32(buf[4:], uint32(len(buf)))
		nativeEndian.PutUint32(buf[8:], uint32(value.Parameter))
		nativeEndian.PutUint32(buf[12:], uint32(1208))
		nativeEndian.PutUint32(buf[16:], uint32(len(value.Strings)))
		nativeEndian.PutUint32(buf[20:], uint32(stringLength))
		offset := stringListBaseLength
		for _, item := range value.Strings {
			copy(buf[offset:offset+stringLength], []byte(item))
			offset += stringLength
		}
		return buf, nil
	case mqconst.MQCFT_BYTE_STRING:
		if len(value.Strings) != 1 {
			return nil, errors.New("byte string values require exactly one base64 string")
		}
		decoded, err := base64.StdEncoding.DecodeString(value.Strings[0])
		if err != nil {
			return nil, fmt.Errorf("decode base64 byte string: %w", err)
		}
		length := roundTo4(len(decoded))
		buf := make([]byte, 16+length)
		nativeEndian.PutUint32(buf[0:], uint32(mqconst.MQCFT_BYTE_STRING))
		nativeEndian.PutUint32(buf[4:], uint32(len(buf)))
		nativeEndian.PutUint32(buf[8:], uint32(value.Parameter))
		nativeEndian.PutUint32(buf[12:], uint32(len(decoded)))
		copy(buf[16:], decoded)
		return buf, nil
	default:
		return nil, fmt.Errorf("unsupported PCF type %d", value.Type)
	}
}

func parseValue(payload []byte) (Value, int, error) {
	if len(payload) < 8 {
		return Value{}, 0, errors.New("payload too short for PCF parameter")
	}

	valueType := int32(nativeEndian.Uint32(payload[0:]))
	strucLength := int(nativeEndian.Uint32(payload[4:]))
	if strucLength > len(payload) || strucLength < 8 {
		return Value{}, 0, fmt.Errorf("invalid PCF parameter length %d", strucLength)
	}

	value := Value{Type: valueType}
	switch valueType {
	case mqconst.MQCFT_INTEGER:
		value.Parameter = int32(nativeEndian.Uint32(payload[8:]))
		value.Integers = []int64{int64(int32(nativeEndian.Uint32(payload[12:])))}
	case mqconst.MQCFT_INTEGER_LIST:
		value.Parameter = int32(nativeEndian.Uint32(payload[8:]))
		count := int(nativeEndian.Uint32(payload[12:]))
		offset := integerListBaseLength
		value.Integers = make([]int64, 0, count)
		for i := 0; i < count; i++ {
			value.Integers = append(value.Integers, int64(int32(nativeEndian.Uint32(payload[offset:]))))
			offset += 4
		}
	case mqconst.MQCFT_STRING:
		value.Parameter = int32(nativeEndian.Uint32(payload[8:]))
		length := int(nativeEndian.Uint32(payload[16:]))
		start := stringBaseLength
		value.Strings = []string{string(bytes.TrimRight(payload[start:start+length], "\x00 "))}
	case mqconst.MQCFT_STRING_LIST:
		value.Parameter = int32(nativeEndian.Uint32(payload[8:]))
		count := int(nativeEndian.Uint32(payload[16:]))
		strLen := int(nativeEndian.Uint32(payload[20:]))
		offset := stringListBaseLength
		value.Strings = make([]string, 0, count)
		for i := 0; i < count; i++ {
			chunk := payload[offset : offset+strLen]
			value.Strings = append(value.Strings, string(bytes.TrimRight(chunk, "\x00 ")))
			offset += strLen
		}
	default:
		return Value{}, 0, fmt.Errorf("unsupported PCF type %d", valueType)
	}

	return value, strucLength, nil
}

func (p stringParameter) encode() ([]byte, error) {
	return encodeValue(Value{
		Type:      mqconst.MQCFT_STRING,
		Parameter: p.parameter,
		Strings:   []string{p.value},
	})
}

func (p integerParameter) encode() ([]byte, error) {
	return encodeValue(Value{
		Type:      mqconst.MQCFT_INTEGER,
		Parameter: p.parameter,
		Integers:  []int64{int64(p.value)},
	})
}

func (p integerListParameter) encode() ([]byte, error) {
	values := make([]int64, 0, len(p.values))
	for _, value := range p.values {
		values = append(values, int64(value))
	}
	return encodeValue(Value{
		Type:      mqconst.MQCFT_INTEGER_LIST,
		Parameter: p.parameter,
		Integers:  values,
	})
}

func roundTo4(length int) int {
	return length + ((4 - (length % 4)) % 4)
}
