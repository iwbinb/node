package communication

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type customMessage struct {
	Field int
}

func TestJsonPack(t *testing.T) {
	packer := JsonPacker(&customMessage{Field: 123})
	data, err := packer.Pack()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"Field": 123}`, string(data))
}

func TestJsonUnpack(t *testing.T) {
	unpacker := JsonPayload{&customMessage{}}
	err := unpacker.Unpack([]byte(`{"Field": 123}`))

	assert.NoError(t, err)
	assert.Equal(t, &customMessage{Field: 123}, unpacker.Model)
}

func TestJsonListener(t *testing.T) {
	var messageConsumed customMessage
	listener := JsonListener(func(message customMessage) {
		messageConsumed = message
	})

	err := listener.Message.Unpack([]byte(`{"Field": 123}`))
	listener.Invoke()

	assert.NoError(t, err)
	assert.Exactly(t, customMessage{123}, messageConsumed)
}

type customRequest struct {
	FieldIn string
}

type customResponse struct {
	FieldOut string
}

func TestJsonHandler(t *testing.T) {
	var requestReceived customRequest
	handler := JsonHandler(func(request customRequest) customResponse {
		requestReceived = request
		return customResponse{"RESPONSE"}
	})

	err := handler.Request.Unpack([]byte(`{"FieldIn": "REQUEST"}`))
	response := handler.Invoke()

	assert.NoError(t, err)
	assert.Exactly(t, customRequest{"REQUEST"}, requestReceived)
	assert.Exactly(t, JsonPayload{customResponse{"RESPONSE"}}, response)
}
