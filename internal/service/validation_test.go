package service

import "testing"

func TestConnectionNormalizeAppliesDefaults(t *testing.T) {
	t.Parallel()

	connection := ConnectionParams{
		Host:         "mq.example.com",
		Port:         1414,
		Channel:      "SYSTEM.ADMIN.SVRCONN",
		QueueManager: "QM1",
		User:         "app",
		Password:     "secret",
	}

	normalized, err := connection.Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if got, want := normalized.ReplyModelQueue, DefaultReplyModelQueue; got != want {
		t.Fatalf("ReplyModelQueue = %q, want %q", got, want)
	}
}

func TestConnectionNormalizeRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	connection := ConnectionParams{
		Host:         "",
		Port:         0,
		Channel:      "",
		QueueManager: "",
		User:         "",
		Password:     "",
	}

	if _, err := connection.Normalize(); err == nil {
		t.Fatal("Normalize() expected an error for missing required fields")
	}
}

func TestPutMessageInputValidateRequiresExactlyOnePayload(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	input := PutTestMessageInput{
		Connection:    connection,
		QueueName:     "DEV.QUEUE.1",
		PayloadText:   "hello",
		PayloadBase64: "aGVsbG8=",
	}

	if err := input.Validate(); err == nil {
		t.Fatal("Validate() expected an error when both payloadText and payloadBase64 are set")
	}
}

func TestPutMessageInputValidateAcceptsTextPayload(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	input := PutTestMessageInput{
		Connection:  connection,
		QueueName:   "DEV.QUEUE.1",
		PayloadText: "hello",
	}

	if err := input.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestUpdateQueueInputValidateRequiresAtLeastOneChange(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	input := UpdateQueueInput{
		Connection: connection,
		QueueName:  "DEV.QUEUE.1",
	}

	if err := input.Validate(); err == nil {
		t.Fatal("Validate() expected an error when no queue updates are requested")
	}
}

func TestCreateChannelInputValidateByType(t *testing.T) {
	t.Parallel()

	connection, err := validConnection().Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	input := CreateChannelInput{
		Connection:  connection,
		ChannelName: "TO.REMOTE.QM",
		ChannelType: "SDR",
	}

	if err := input.Validate(); err == nil {
		t.Fatal("Validate() expected an error when SDR channel has no connectionName/xmitQueueName")
	}

	input.ConnectionName = "remote.example.com(1414)"
	input.XmitQueueName = "QM2.XMITQ"

	if err := input.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDecodePayloadPrefersTextAndFallsBackToBase64(t *testing.T) {
	t.Parallel()

	textPreview := DecodePayload([]byte("hello"))
	if textPreview.Format != "text" || textPreview.Text != "hello" {
		t.Fatalf("DecodePayload(text) = %#v, want text preview", textPreview)
	}

	binaryPreview := DecodePayload([]byte{0xff, 0x00, 0xfe})
	if binaryPreview.Format != "base64" || binaryPreview.Base64 == "" {
		t.Fatalf("DecodePayload(binary) = %#v, want base64 preview", binaryPreview)
	}
}

func validConnection() ConnectionParams {
	return ConnectionParams{
		Host:         "mq.example.com",
		Port:         1414,
		Channel:      "SYSTEM.ADMIN.SVRCONN",
		QueueManager: "QM1",
		User:         "app",
		Password:     "secret",
	}
}
