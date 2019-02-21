package bootstrap

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/message"
	"github.com/golemfactory/bootstrap_go/python"
	"golang.org/x/crypto/sha3"
)

type PeerSession struct {
	service *Service
	conn    net.Conn
	pubKey  crypto.PublicKey
	inited  bool
	peer    python.Peer
	id      string
}

func NewPeerSession(service *Service, conn net.Conn) *PeerSession {
	return &PeerSession{
		service: service,
		conn:    conn,
	}
}

func (session *PeerSession) Close() {
	session.conn.Close()
}

func (session *PeerSession) sendDisconnect(reason message.DisconnectReason) error {
	return session.sendMessage(&message.Disconnect{Reason: reason})
}

func (session *PeerSession) performHandshake() error {
	conn := session.conn
	service := session.service

	myHello := service.genHello()
	err := session.sendMessage(myHello)
	if err != nil {
		return fmt.Errorf("send hello error: %v", err)
	}
	msg, err := session.receiveMessage()
	if err != nil {
		return fmt.Errorf("receive hello error: %v", err)
	}
	if disconnectMsg, ok := msg.(*message.Disconnect); ok {
		return fmt.Errorf("peer disconnected, reason: %v", disconnectMsg.Reason)
	}

	helloMsg, ok := msg.(*message.Hello)
	if !ok {
		return fmt.Errorf("was expecting Hello, got type %d", msg.GetType())
	}

	if helloMsg.ProtoId != session.service.config.ProtocolId {
		if err := session.sendDisconnect(message.DISCONNECT_PROTOCOL_VERSION); err != nil {
			return err
		}
		return fmt.Errorf("not matching protocol ID, remote %v, local %v", helloMsg.ProtoId, session.service.config.ProtocolId)
	}

	nodeInfo, err := python.DictToNode(helloMsg.NodeInfo)
	if err != nil {
		return fmt.Errorf("Malformed node info: %v", err)
	}

	pubKeyBytes, err := hex.DecodeString(nodeInfo.Key)
	if err != nil {
		return fmt.Errorf("couldn't decode remote public key: %v", err)
	}
	session.pubKey, err = crypto.PublicKeyFromBytes(append([]byte{0x04}, pubKeyBytes...))
	if err != nil {
		return fmt.Errorf("couldn't create remote public key: %v", err)
	}
	keyDifficulty := crypto.GetKeyDifficulty(session.pubKey)
	if keyDifficulty < session.service.config.KeyDifficulty {
		if err := session.sendDisconnect(message.DISCONNECT_KEY_DIFFICULTY); err != nil {
			return err
		}
		return fmt.Errorf("key not difficult enough, got %v", keyDifficulty)
	}
	session.inited = true

	msg, err = session.receiveMessage()
	if err != nil {
		return fmt.Errorf("receive randval error: %v", err)
	}
	if disconnectMsg, ok := msg.(*message.Disconnect); ok {
		return fmt.Errorf("peer disconnected, reason: %v", disconnectMsg.Reason)
	}

	randValMsg, ok := msg.(*message.RandVal)
	if !ok {
		return fmt.Errorf("expected RandVal message, got type %d", msg.GetType())
	}
	if randValMsg.RandVal != myHello.RandVal {
		return fmt.Errorf("incorrect RandVal value")
	}

	myRandValMsg := message.RandVal{RandVal: helloMsg.RandVal}
	err = session.sendMessage(&myRandValMsg)
	if err != nil {
		return err
	}

	addr, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	session.peer = python.Peer{
		Address:  addr,
		Port:     helloMsg.Port,
		Node:     nodeInfo,
		NodeName: helloMsg.NodeName,
	}
	session.id = helloMsg.ClientKeyId
	return nil
}

func (session *PeerSession) handle() error {
	err := session.performHandshake()
	if err != nil {
		return err
	}

	pk := session.service.peerKeeper
	peers := pk.GetPeers(session.id)
	peersMsg := &message.Peers{
		Peers: make([]interface{}, len(peers)),
	}
	for idx, p := range peers {
		peersMsg.Peers[idx] = p.ToDict()
	}
	err = session.sendMessage(peersMsg)
	if err != nil {
		return err
	}
	pk.AddPeer(session.id, session.peer)

	disconnectMsg := &message.Disconnect{
		Reason: message.DISCONNECT_BOOTSTRAP,
	}
	err = session.sendMessage(disconnectMsg)
	if err != nil {
		return err
	}

	return nil
}

func (session *PeerSession) receiveMessage() (message.Message, error) {
	return message.Receive(session.conn, session.decrypt, session.verifySign)
}

func (session *PeerSession) sendMessage(msg message.Message) error {
	return message.Send(
		session.conn,
		msg,
		session.encrypt,
		session.sign)
}

func (session *PeerSession) decrypt(data []byte) ([]byte, error) {
	res, err := session.service.privKey.Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt message: %v", err)
	}
	return res, nil
}

func (session *PeerSession) encrypt(data []byte) ([]byte, error) {
	res, err := crypto.Encrypt(data, session.pubKey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt message: %v", err)
	}
	return res, nil
}

func GetShortHashSha(data []byte) []byte {
	sha := sha3.New256()
	sha.Write(data)
	return sha.Sum(nil)
}

func (session *PeerSession) sign(shortHash []byte) ([]byte, error) {
	return session.service.privKey.Sign(GetShortHashSha(shortHash))
}

func (session *PeerSession) verifySign(shortHash []byte, sig []byte) bool {
	if !session.inited {
		return true
	}
	return session.pubKey.VerifySign(GetShortHashSha(shortHash), sig)
}
