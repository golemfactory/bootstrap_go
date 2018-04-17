package bootstrap

import (
	"net"
	"testing"
	"time"

	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/message"
	"github.com/golemfactory/bootstrap_go/peerkeeper"
	"github.com/golemfactory/bootstrap_go/python"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TEST_NAME     = "bootstrap-unittest"
	TEST_PROTO_ID = "1337"
)

type TestAddress struct {
}

func (a *TestAddress) Network() string {
	return "test-network"
}

func (a *TestAddress) String() string {
	return "test-addr:test-port"
}

type TestConn struct {
	net.Conn
}

func (c *TestConn) RemoteAddr() net.Addr {
	return &TestAddress{}
}

type AddPeerCall struct {
	Id   string
	Peer python.Peer
}

type GetPeersCall struct {
	Id string
}

type TestPeerKeeper struct {
	AddPeerCalls  []AddPeerCall
	GetPeersCalls []GetPeersCall
}

func NewTestPeerKeeper() *TestPeerKeeper {
	return &TestPeerKeeper{
		AddPeerCalls:  make([]AddPeerCall, 0),
		GetPeersCalls: make([]GetPeersCall, 0),
	}
}

func (pk *TestPeerKeeper) AddPeer(id string, peer python.Peer) {
	pk.AddPeerCalls = append(pk.AddPeerCalls, AddPeerCall{id, peer})
}

func (pk *TestPeerKeeper) GetPeers(id string) []python.Peer {
	pk.GetPeersCalls = append(pk.GetPeersCalls, GetPeersCall{id})
	return nil
}

func getService(t *testing.T, pk peerkeeper.PeerKeeper, keyDifficulty int) *Service {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fatal("Error while generating private key", err)
	}

	config := &Config{
		Name:          TEST_NAME,
		Id:            "deadbeef",
		Port:          44444,
		PrvAddr:       "prvAddr",
		PubAddr:       "pubAddr",
		PrvAddresses:  nil,
		NatType:       "nat type",
		PeerNum:       100,
		KeyDifficulty: keyDifficulty,
		ProtocolId:    TEST_PROTO_ID,
	}

	return NewService(config, privKey, pk)
}

func testPeerSessionImpl(t *testing.T, handleCh chan error) {
	const (
		RAND_VAL  = 0.1337
		CLIENT_ID = "client-id"
	)
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	pubKey := privKey.GetPublicKey()
	pubKeyHex := pubKey.Hex()

	pk := &TestPeerKeeper{}
	service := getService(t, pk, 0)
	conn, psConn := net.Pipe()
	ps := NewPeerSession(service, &TestConn{Conn: psConn})
	go func() {
		handleCh <- ps.handle()
	}()

	signFunc := func(msg message.Message) ([]byte, error) {
		return privKey.Sign(GetShortHashSha(msg))
	}
	encryptFunc := func(data []byte) ([]byte, error) {
		return crypto.Encrypt(data, service.privKey.GetPublicKey())
	}
	decryptFunc := func(data []byte) ([]byte, error) {
		return privKey.Decrypt(data)
	}

	msg, err := message.Receive(conn, nil)
	require.NoError(t, err)
	serverHello := msg.(*message.Hello)
	assert.Equal(t, TEST_NAME, serverHello.NodeName)

	node := python.Node{
		Key: pubKeyHex,
	}
	hello := &message.Hello{
		RandVal:     RAND_VAL,
		ClientKeyId: CLIENT_ID,
		NodeInfo:    node.ToDict(),
		ProtoId:     TEST_PROTO_ID,
	}
	err = message.Send(conn, hello, encryptFunc, signFunc)
	require.NoError(t, err)

	randVal := &message.RandVal{
		RandVal: serverHello.RandVal,
	}
	err = message.Send(conn, randVal, encryptFunc, signFunc)
	require.NoError(t, err)

	msg, err = message.Receive(conn, decryptFunc)
	require.NoError(t, err)
	serverRandVal := msg.(*message.RandVal)
	assert.Equal(t, RAND_VAL, serverRandVal.RandVal)

	msg, err = message.Receive(conn, decryptFunc)
	require.NoError(t, err)
	serverPeers := msg.(*message.Peers)
	assert.Equal(t, 0, len(serverPeers.Peers))

	msg, err = message.Receive(conn, decryptFunc)
	require.NoError(t, err)
	disconnect := msg.(*message.Disconnect)
	assert.Equal(t, message.DISCONNECT_BOOTSTRAP, disconnect.Reason)

	require.Equal(t, 1, len(pk.GetPeersCalls))
	assert.Equal(t, CLIENT_ID, pk.GetPeersCalls[0].Id)
	require.Equal(t, 1, len(pk.AddPeerCalls))
	assert.Equal(t, CLIENT_ID, pk.AddPeerCalls[0].Id)
}

func TestPeerSession(t *testing.T) {
	testCh := make(chan bool)
	handleCh := make(chan error)
	go func() {
		testPeerSessionImpl(t, handleCh)
		close(testCh)
	}()

	select {
	case <-testCh:
	case err := <-handleCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Test timed out")
	}
}

func TestDisconnectKeyDifficulty(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	pubKey := privKey.GetPublicKey()
	pubKeyHex := pubKey.Hex()

	pk := &TestPeerKeeper{}
	service := getService(t, pk, 100)
	conn, psConn := net.Pipe()
	ps := NewPeerSession(service, &TestConn{Conn: psConn})
	go func() {
		ps.handle()
	}()

	signFunc := func(msg message.Message) ([]byte, error) {
		return privKey.Sign(GetShortHashSha(msg))
	}

	msg, err := message.Receive(conn, nil)
	require.NoError(t, err)

	node := python.Node{
		Key: pubKeyHex,
	}
	hello := &message.Hello{
		NodeInfo: node.ToDict(),
		ProtoId:  TEST_PROTO_ID,
	}
	err = message.Send(conn, hello, nil, signFunc)
	require.NoError(t, err)

	msg, err = message.Receive(conn, nil)
	require.NoError(t, err)

	require.Equal(t, message.MSG_DISCONNECT_TYPE, int(msg.GetType()))
	disconnectMsg := msg.(*message.Disconnect)
	assert.Equal(t, message.DISCONNECT_KEY_DIFFICULTY, disconnectMsg.Reason)
}
