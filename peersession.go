package bootstrap

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/message"
	"github.com/golemfactory/bootstrap_go/python"
	"golang.org/x/crypto/sha3"
)

type PeerSession struct {
	service *Service
	conn    net.Conn
	pubKey  crypto.PublicKey
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
		return err
	}
	msg, err := session.receiveMessage()
	if err != nil {
		return err
	}
	if msg.GetType() == message.MSG_DISCONNECT_TYPE {
		if disconnectMsg, ok := msg.(*message.Disconnect); ok {
			return fmt.Errorf("peer disconnected, reason: %v", disconnectMsg.Reason)
		}
		return fmt.Errorf("wrong message type, was expecting Disconnect")
	}
	if msg.GetType() != message.MSG_HELLO_TYPE {
		return fmt.Errorf("unexpected msg type %d, was expecting Hello", msg.GetType())
	}

	helloMsg := msg.(*message.Hello)

	if helloMsg.ProtoId != session.service.config.ProtocolId {
		if err := session.sendDisconnect(message.DISCONNECT_PROTOCOL_VERSION); err != nil {
			return err
		}
		return fmt.Errorf("not matching protocol ID, remote %v, local %v", helloMsg.ProtoId, session.service.config.ProtocolId)
	}

	nodeInfo := python.DictToNode(helloMsg.NodeInfo)

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

	msg, err = session.receiveMessage()
	if err != nil {
		return err
	}
	if msg.GetType() == message.MSG_DISCONNECT_TYPE {
		if disconnectMsg, ok := msg.(*message.Disconnect); ok {
			return fmt.Errorf("peer disconnected, reason: %v", disconnectMsg.Reason)
		}
		return fmt.Errorf("wrong message type, was expecting Disconnect")
	}
	if msg.GetType() != message.MSG_RAND_VAL_TYPE {
		return fmt.Errorf("unexpected msg type %d, was expecting RandVal", msg.GetType())
	}
	randValMsg := msg.(*message.RandVal)
	if randValMsg.RandVal != myHello.RandVal {
		return fmt.Errorf("incorrect RandVal value")
	}

	signed, err := session.verifySign(randValMsg)
	if !signed || err != nil {
		if err := session.sendDisconnect(message.DISCONNECT_UNVERIFIED); err != nil {
			return err
		}
		return fmt.Errorf("RandVal message not signed correctly")
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
	fmt.Println("Peer connection from", session.conn.RemoteAddr())
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
	return message.Receive(session.conn, session.decrypt)
}

func (session *PeerSession) sendMessage(msg message.Message) error {
	return message.Send(
		session.conn,
		msg,
		func(data []byte) ([]byte, error) {
			return session.encrypt(data)
		},
		session.sign)
}

func (session *PeerSession) decrypt(data []byte) ([]byte, error) {
	res, err := crypto.DecryptPython(session.service.privKey, data)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt message: %v", err)
	}
	return res, nil
}

func (session *PeerSession) encrypt(data []byte) ([]byte, error) {
	res, err := crypto.EncryptPython(session.service.privKey, data, session.pubKey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt message: %v", err)
	}
	return res, nil
}

func GetShortHashSha(msg message.Message) []byte {
	data := msg.GetShortHash(message.GetPayload(msg))
	sha := sha3.New256()
	sha.Write(data)
	return sha.Sum(nil)
}

func (session *PeerSession) sign(msg message.Message) {
	sig, _ := secp256k1.Sign(GetShortHashSha(msg), session.service.privKey.Key)
	msg.GetBaseMessage().Sig = sig
}

func (session *PeerSession) verifySign(msg message.Message) (bool, error) {
	keyBytes := []byte{0x04}
	keyBytes = append(keyBytes, session.pubKey.X...)
	keyBytes = append(keyBytes, session.pubKey.Y...)
	recoveredKey, err := secp256k1.RecoverPubkey(GetShortHashSha(msg), msg.GetBaseMessage().Sig)
	if err != nil {
		return false, fmt.Errorf("unable to recover public key: %v", err)
	}
	return bytes.Equal(recoveredKey, keyBytes), nil
}
