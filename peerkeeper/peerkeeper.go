package peerkeeper

import "github.com/golemfactory/bootstrap_go/python"

// Implementations should be thread safe.
type PeerKeeper interface {
	AddPeer(id string, peer python.Peer)
	GetPeers(id string) []python.Peer
}
