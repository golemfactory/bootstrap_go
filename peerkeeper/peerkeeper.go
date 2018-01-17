package peerkeeper

import (
	"github.com/golemfactory/bootstrap_go/python"
)

type Peer struct {
	Address  string       `cbor:"address"`
	Port     uint64       `cbor:"port"`
	Node     *python.Node `cbor:"node"`
	NodeName string       `cbor:"node_name"`
}

// Implementations should be thread safe.
type PeerKeeper interface {
	AddPeer(id string, peer Peer)
	GetPeers(id string) []Peer
}
