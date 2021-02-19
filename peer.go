package protoo

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

type sentInfo struct {
	id     uint32
	method string
	respCh chan PeerResponse
}

type PeerResponse struct {
	data json.RawMessage
	err  error
}

func (r PeerResponse) Unmarshal(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(r.data) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(r.data), v)
}

func (r PeerResponse) Data() []byte {
	return []byte(r.data)
}

func (r PeerResponse) Err() error {
	return r.err
}

type Peer struct {
	IEventEmitter
	logger    logr.Logger
	locker    sync.Mutex
	id        string
	transport Transport
	sents     map[uint32]sentInfo
	data      interface{}
	closed    bool
	closeCh   chan struct{}
}

func NewPeer(peerId string, data interface{}, transport Transport) *Peer {
	peer := &Peer{
		IEventEmitter: NewEventEmitter(),
		logger:        NewLogger("Peer"),
		id:            peerId,
		transport:     transport,
		sents:         make(map[uint32]sentInfo),
		data:          data,
		closeCh:       make(chan struct{}),
	}
	peer.handleTransport()

	return peer
}

func (peer *Peer) Id() string {
	return peer.id
}

func (peer *Peer) Data() interface{} {
	return peer.data
}

func (peer *Peer) Close() {
	peer.locker.Lock()
	defer peer.locker.Unlock()

	if peer.closed {
		return
	}

	peer.closed = true
	close(peer.closeCh)
	peer.SafeEmit("close")
}

func (peer *Peer) Request(method string, data interface{}) (rsp PeerResponse) {
	request := CreateRequest(method, data)

	sent := sentInfo{
		id:     request.Id,
		method: method,
		respCh: make(chan PeerResponse),
	}

	peer.locker.Lock()

	size := len(peer.sents)
	peer.sents[sent.id] = sent

	peer.locker.Unlock()

	defer func() {
		peer.locker.Lock()

		delete(peer.sents, sent.id)

		peer.locker.Unlock()
	}()

	if err := peer.transport.Send(request.Marshal()); err != nil {
		rsp.err = err
		return
	}

	timeout := 2000 * (15 + (0.1 * float64(size)))
	timer := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	defer timer.Stop()

	select {
	case rsp = <-sent.respCh:
	case <-timer.C:
		rsp.err = errors.New("request timeout")
	case <-peer.closeCh:
		rsp.err = errors.New("peer closed")
	}

	return
}

func (peer *Peer) Notify(method string, data interface{}) error {
	notification := CreateNotification(method, data)

	return peer.transport.Send(notification.Marshal())
}

func (peer *Peer) handleTransport() {
	if peer.transport.Closed() {
		peer.Close()
		return
	}

	peer.transport.On("close", func() {
		peer.Close()
	})

	peer.transport.On("message", func(message Message) {
		if message.Request {
			peer.handleRequest(message)
		} else if message.Response {
			peer.handleResponse(message)
		} else if message.Notification {
			peer.handleNotification(message)
		}
	})
}

func (peer *Peer) handleRequest(request Message) {
	peer.SafeEmit("request", request, func(data interface{}) {
		response := CreateSuccessResponse(request, data)
		peer.transport.Send(response.Marshal())
	}, func(err error) {
		var anErr *Error
		e1, ok := err.(*Error)
		if ok {
			anErr = e1
		} else if e2, ok := err.(Error); ok {
			anErr = &e2
		} else {
			anErr = NewError(500, err.Error())
		}
		response := CreateErrorResponse(request, anErr)
		peer.transport.Send(response.Marshal())
	})
}

func (peer *Peer) handleResponse(response Message) {
	peer.locker.Lock()

	sent, ok := peer.sents[response.Id]

	if !ok {
		peer.locker.Unlock()
		err := errors.New("bad response")
		peer.logger.Error(err, "received response does not match any sent request", "id", response.Id)
		return
	}

	delete(peer.sents, response.Id)

	peer.locker.Unlock()

	if response.OK {
		sent.respCh <- PeerResponse{
			data: response.Data,
		}
	} else {
		sent.respCh <- PeerResponse{
			err: response.Error,
		}
	}
}

func (peer *Peer) handleNotification(notification Message) {
	peer.SafeEmit("notification", notification)
}
