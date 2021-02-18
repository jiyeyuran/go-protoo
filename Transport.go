package protoo

import (
	"fmt"

	"github.com/jiyeyuran/go-eventemitter"
)

type Transport interface {
	eventemitter.IEventEmitter
	fmt.Stringer
	Send(data []byte) error
	Close()
	Closed() bool
	Run() error
}
