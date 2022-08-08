package transport

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/jiyeyuran/go-eventemitter"
	"github.com/jiyeyuran/go-protoo"
)

type WebsocketTransport struct {
	eventemitter.IEventEmitter
	logger logr.Logger
	mu     sync.Mutex
	conn   *websocket.Conn
	closed uint32
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
	t.mu.Lock()
	defer t.mu.Unlock()
	err := t.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		t.Close()
	}
	return err
}

func (t *WebsocketTransport) Close() {
	if atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
		t.logger.V(1).Info("close()", "conn", t.String())
		t.conn.Close()
		t.SafeEmit("close")
		t.RemoveAllListeners()
	}
}

func (t *WebsocketTransport) Closed() bool {
	return atomic.LoadUint32(&t.closed) > 0
}

func (t *WebsocketTransport) String() string {
	return t.conn.RemoteAddr().String()
}

func (t *WebsocketTransport) Run() error {
	for {
		messageType, data, err := t.conn.ReadMessage()

		if err != nil {
			t.Close()
			return err
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

		t.Emit("message", message)
	}
}
