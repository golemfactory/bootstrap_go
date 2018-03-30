package peerkeeper

import (
	"sync"

	"github.com/golemfactory/bootstrap_go/python"
)

type RandomizedPeerKeeper struct {
	peers   map[string]python.Peer
	peerNum int
	mutex   sync.Mutex
}

func NewRandomizedPeerKeeper(peerNum int) *RandomizedPeerKeeper {
	return &RandomizedPeerKeeper{
		peers:   make(map[string]python.Peer),
		peerNum: peerNum,
		mutex:   sync.Mutex{},
	}
}

func (pk *RandomizedPeerKeeper) AddPeer(id string, peer python.Peer) {
	pk.mutex.Lock()
	defer pk.mutex.Unlock()
	if _, ok := pk.peers[id]; ok {
		return
	}
	if len(pk.peers) >= pk.peerNum {
		// remove a random peer and since map iteration order is random
		// we can remove the first peer we encounter
		for id, _ := range pk.peers {
			delete(pk.peers, id)
			break
		}
	}
	pk.peers[id] = peer
}

func (pk *RandomizedPeerKeeper) GetPeers(peerId string) []python.Peer {
	pk.mutex.Lock()
	defer pk.mutex.Unlock()
	peers := make([]python.Peer, 0, len(pk.peers))
	for id, p := range pk.peers {
		if id != peerId {
			peers = append(peers, p)
		}
	}
	return peers
}
