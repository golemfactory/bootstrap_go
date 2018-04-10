package peerkeeper

import (
	"testing"

	"github.com/golemfactory/bootstrap_go/python"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomizedPeerKeeper(t *testing.T) {
	pk := NewRandomizedPeerKeeper(2)
	peers := pk.GetPeers("foo")
	require.Equal(t, 0, len(peers))

	peer1 := python.Peer{NodeName: "peer1"}
	pk.AddPeer("peer1", peer1)
	peers = pk.GetPeers("foo")
	require.Equal(t, 1, len(peers))
	assert.Equal(t, "peer1", peers[0].NodeName)

	peer2 := python.Peer{NodeName: "peer2"}
	pk.AddPeer("peer2", peer2)
	peers = pk.GetPeers("foo")
	require.Equal(t, 2, len(peers))

	peers = pk.GetPeers("peer2")
	require.Equal(t, 1, len(peers))
	assert.Equal(t, "peer1", peers[0].NodeName)

	peer3 := python.Peer{NodeName: "peer3"}
	pk.AddPeer("peer3", peer3)
	peers = pk.GetPeers("foo")
	require.Equal(t, 2, len(peers))
	if peers[0].NodeName != "peer3" && peers[1].NodeName != "peer3" {
		t.Errorf("Expected peer3 to be in the list, got %v", peers)
	}
}
