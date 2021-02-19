package protoo

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	OK           bool            `json:"ok,omitempty"`
	Request      bool            `json:"request,omitempty"`
	Response     bool            `json:"response,omitempty"`
	Notification bool            `json:"notification,omitempty"`
	Id           uint32          `json:"id,omitempty"`
	Method       string          `json:"method,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`

	*Error
}

type Error struct {
	ErrorCode   int    `json:"errorCode,omitempty"`
	ErrorReason string `json:"errorReason,omitempty"`
}

func NewError(code int, reason string) *Error {
	return &Error{
		ErrorCode:   code,
		ErrorReason: reason,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.ErrorCode, e.ErrorReason)
}

func (m Message) String() string {
	data, _ := json.Marshal(m)

	return string(data)
}

func (m Message) Marshal() []byte {
	data, _ := json.Marshal(m)

	return data
}

func CreateRequest(method string, data interface{}) Message {
	raw, _ := json.Marshal(data)

	return Message{
		Request: true,
		Id:      generateRandomNumber(),
		Method:  method,
		Data:    json.RawMessage(raw),
	}
}

func CreateSuccessResponse(request Message, data interface{}) Message {
	raw, _ := json.Marshal(data)

	return Message{
		Response: true,
		Id:       request.Id,
		OK:       true,
		Data:     json.RawMessage(raw),
	}
}

func CreateErrorResponse(request Message, err *Error) Message {
	return Message{
		Response: true,
		Id:       request.Id,
		Error:    err,
	}
}

func CreateNotification(method string, data interface{}) Message {
	raw, _ := json.Marshal(data)

	return Message{
		Notification: true,
		Method:       method,
		Data:         json.RawMessage(raw),
	}
}
