package protoo

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
)

type Room struct {
	IEventEmitter
	locker sync.Mutex
	logger logr.Logger
	closed bool
	peers  map[string]*Peer
}

func NewRoom() *Room {
	return &Room{
		IEventEmitter: NewEventEmitter(),
		logger:        NewLogger("Room"),
		peers:         make(map[string]*Peer),
	}
}

func (r *Room) Closed() bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	return r.closed
}

func (r *Room) Peers() (peers []*Peer) {
	r.locker.Lock()
	defer r.locker.Unlock()

	for _, peer := range r.peers {
		peers = append(peers, peer)
	}

	return
}

func (r *Room) Close() {
	r.locker.Lock()
	defer r.locker.Unlock()

	r.logger.V(1).Info("close()")

	for _, peer := range r.peers {
		peer.Close()
	}

	r.peers = make(map[string]*Peer)
	r.SafeEmit("close")
	r.RemoveAllListeners()
}

func (r *Room) CreatePeer(peerId string, peerData interface{}, transport Transport) (peer *Peer, err error) {
	r.locker.Lock()
	defer r.locker.Unlock()

	r.logger.V(1).Info("createPeer()", "peerId", peerId, "transport", transport.String())

	if _, ok := r.peers[peerId]; ok {
		transport.Close()
		err = fmt.Errorf(`there is already a Peer with same peerId [peerId:"%s"]`, peerId)
		return
	}

	peer = NewPeer(peerId, peerData, transport)
	r.peers[peerId] = peer

	peer.On("close", func() {
		r.locker.Lock()
		defer r.locker.Unlock()
		delete(r.peers, peerId)
	})

	return
}

func (r *Room) HasPeer(peerId string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	_, ok := r.peers[peerId]

	return ok
}

func (r *Room) GetPeer(peerId string) *Peer {
	r.locker.Lock()
	defer r.locker.Unlock()

	return r.peers[peerId]
}
