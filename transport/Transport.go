package transport

import "github.com/jiyeyuran/go-eventemitter"

type Transport interface {
	eventemitter.IEventEmitter
	Id() string
	Send(data []byte) error
	Close()
	Closed() bool
}
