package transport

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/jiyeyuran/go-eventemitter"
	"github.com/jiyeyuran/go-protoo"
)

type WebsocketTransport struct {
	eventemitter.IEventEmitter
	logger logr.Logger
	locker sync.Mutex
	conn   *websocket.Conn
	closed bool
}

func NewWebsocketTransport(conn *websocket.Conn) protoo.Transport {
	t := &WebsocketTransport{
		IEventEmitter: protoo.NewEventEmitter(),
		logger:        protoo.NewLogger("WebSocketTransport"),
		conn:          conn,
	}

	return t
}

func (t *WebsocketTransport) Send(message []byte) error {
	if t.Closed() {
		return errors.New("transport closed")
	}
	return t.conn.WriteMessage(websocket.TextMessage, message)
}

func (t *WebsocketTransport) Close() {
	t.locker.Lock()
	defer t.locker.Unlock()

	if t.closed {
		return
	}

	t.logger.V(1).Info("close()", "conn", t.String())

	t.closed = true
	t.conn.Close()
	t.SafeEmit("close")
}

func (t *WebsocketTransport) Closed() bool {
	t.locker.Lock()
	defer t.locker.Unlock()

	return t.closed
}

func (t *WebsocketTransport) String() string {
	return t.conn.RemoteAddr().String()
}

func (t *WebsocketTransport) Run() {
	for {
		messageType, data, err := t.conn.ReadMessage()

		if err != nil {
			t.Close()
			return
		}

		if messageType == websocket.BinaryMessage {
			t.logger.V(0).Info("warning of ignoring received binary message", "conn", t.String())
			continue
		}

		if t.ListenerCount("message") == 0 {
			err := errors.New(`no listeners for "message" event`)
			t.logger.Error(err, `ignoring received message`, "conn", t.String())
			continue
		}

		message := protoo.Message{}

		if err := json.Unmarshal(data, &message); err != nil {
			t.logger.Error(err, `json unmarshal`, "conn", t.String())
			continue
		}

		t.SafeEmit("message", message)
	}
}
