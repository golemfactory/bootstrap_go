package bootstrap

import (
	"net"
	"testing"
	"time"

	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/message"
	"github.com/golemfactory/bootstrap_go/peerkeeper"
	"github.com/golemfactory/bootstrap_go/python"
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
	if err != nil {
		t.Fatal("Error while generating private key", err)
	}
	pubKeyHex := privKey.GetPubKeyHex()

	pk := &TestPeerKeeper{}
	service := getService(t, pk, 0)
	conn, psConn := net.Pipe()
	ps := NewPeerSession(service, &TestConn{Conn: psConn})
	go func() {
		handleCh <- ps.handle()
	}()

	signFunc := func(msg message.Message) {
		sig, _ := privKey.Sign(GetShortHashSha(msg))
		msg.GetBaseMessage().Sig = sig
	}
	encryptFunc := func(data []byte) ([]byte, error) {
		return crypto.Encrypt(data, service.privKey.GetPublicKey())
	}
	decryptFunc := func(data []byte) ([]byte, error) {
		return privKey.Decrypt(data)
	}

	msg, err := message.Receive(conn, nil)
	if err != nil {
		t.Fatal(err)
	}
	serverHello := msg.(*message.Hello)
	if serverHello.NodeName != TEST_NAME {
		t.Error("Wrong bootstrap node name:", serverHello.NodeName)
	}

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
	if err != nil {
		t.Fatal(err)
	}

	randVal := &message.RandVal{
		RandVal: serverHello.RandVal,
	}
	err = message.Send(conn, randVal, encryptFunc, signFunc)
	if err != nil {
		t.Fatal(err)
	}

	msg, err = message.Receive(conn, decryptFunc)
	if err != nil {
		t.Fatal(err)
	}
	serverRandVal := msg.(*message.RandVal)
	if serverRandVal.RandVal != RAND_VAL {
		t.Fatal("Wrong RandVal", serverRandVal.RandVal)
	}

	msg, err = message.Receive(conn, decryptFunc)
	if err != nil {
		t.Fatal(err)
	}
	serverPeers := msg.(*message.Peers)
	if len(serverPeers.Peers) != 0 {
		t.Fatalf("Expected empty list of peers, got %+v", serverPeers.Peers)
	}

	msg, err = message.Receive(conn, decryptFunc)
	if err != nil {
		t.Fatal(err)
	}
	disconnect := msg.(*message.Disconnect)
	if disconnect.Reason != message.DISCONNECT_BOOTSTRAP {
		t.Fatal("Expected disconnect bootstrap, got:", disconnect.Reason)
	}

	if len(pk.GetPeersCalls) != 1 {
		t.Fatal("GetPeers should be called once, was called:", len(pk.GetPeersCalls))
	}
	if pk.GetPeersCalls[0].Id != CLIENT_ID {
		t.Fatal("GetPeers was called with wrong Id:", pk.GetPeersCalls[0].Id)
	}
	if len(pk.AddPeerCalls) != 1 {
		t.Fatal("AddPeer should be called once, was called:", len(pk.AddPeerCalls))
	}
	if pk.AddPeerCalls[0].Id != CLIENT_ID {
		t.Fatal("AddPeer was called with wrong Id:", pk.AddPeerCalls[0].Id)
	}
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
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("Test timed out")
	}
}

func TestDisconnectKeyDifficulty(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fatal("Error while generating private key", err)
	}
	pubKeyHex := privKey.GetPubKeyHex()

	pk := &TestPeerKeeper{}
	service := getService(t, pk, 100)
	conn, psConn := net.Pipe()
	ps := NewPeerSession(service, &TestConn{Conn: psConn})
	go func() {
		ps.handle()
	}()

	signFunc := func(msg message.Message) {
		sig, _ := privKey.Sign(GetShortHashSha(msg))
		msg.GetBaseMessage().Sig = sig
	}

	msg, err := message.Receive(conn, nil)
	if err != nil {
		t.Fatal(err)
	}

	node := python.Node{
		Key: pubKeyHex,
	}
	hello := &message.Hello{
		NodeInfo: node.ToDict(),
		ProtoId:  TEST_PROTO_ID,
	}
	err = message.Send(conn, hello, nil, signFunc)
	if err != nil {
		t.Fatal(err)
	}

	msg, err = message.Receive(conn, nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg.GetType() != message.MSG_DISCONNECT_TYPE {
		t.Fatal("Wrong msg type, expected disconnect, got", msg.GetType())
	}
	disconnectMsg := msg.(*message.Disconnect)
	if disconnectMsg.Reason != "key_not_difficult" {
		t.Error("Expected reason `key_not_difficult`, got", disconnectMsg.Reason)
	}
}
