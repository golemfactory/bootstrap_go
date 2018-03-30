package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/golemfactory/bootstrap_go"
	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/peerkeeper"

	"github.com/ccding/go-stun/stun"
)

const (
	PORT                   = 40102
	PEER_NUM               = 100
	NAME                   = "Go Bootstrap"
	PROTO_ID               = "26"
	GOLEM_MESSAGES_VERSION = "1.17.2"
	GOLEM_VERSION          = "0.14.0"
	KEY_DIFF               = 14
)

func main() {
	var port uint64
	var peerNum int
	var name string
	var protocolId string
	var golemMessagesVersion string
	var golemVersion string
	var mainnet bool
	flag.Uint64Var(&port, "port", PORT, "Port to listen to")
	flag.IntVar(&peerNum, "peer-num", PEER_NUM, "Number of peers to send")
	flag.StringVar(&name, "name", NAME, "Name of the node")
	flag.StringVar(&protocolId, "protocol-id", PROTO_ID, "Version of the P2P procotol")
	flag.StringVar(&golemMessagesVersion, "golem-messages", GOLEM_MESSAGES_VERSION, "Version of the golem-messages library")
	flag.StringVar(&golemVersion, "golem-version", GOLEM_VERSION, "Version of Golem")
	flag.BoolVar(&mainnet, "mainnet", false, "Whether to run on a mainnet")
	flag.Parse()

	if !mainnet {
		protocolId += "-testnet"
	}

	var err error
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return
	}
	prvAddresses := make([]interface{}, 0)
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				prvAddresses = append(prvAddresses, ipnet.IP.String())
			}
		}
	}

	nat, host, err := stun.NewClient().Discover()
	if err != nil {
		fmt.Println("Error discovering STUN details:", err)
		return
	}

	rand.Seed(time.Now().UTC().UnixNano())
	privKey, err := crypto.GenerateDifficultKey(KEY_DIFF)
	if err != nil {
		fmt.Println("Error while generating private key", err)
		return
	}
	pubKeyHex := crypto.GetPubKeyHex(privKey)

	config := &bootstrap.Config{
		Name:                 name,
		Id:                   pubKeyHex,
		Port:                 port,
		PrvAddr:              prvAddresses[0].(string),
		PubAddr:              host.IP(),
		PrvAddresses:         prvAddresses,
		NatType:              nat.String(),
		PeerNum:              PEER_NUM,
		ProtocolId:           protocolId,
		GolemMessagesVersion: golemMessagesVersion,
		GolemVersion:         golemVersion,
	}

	fmt.Printf("Config: %+v\n", config)

	service := bootstrap.NewService(
		config,
		privKey,
		peerkeeper.NewRandomizedPeerKeeper(config.PeerNum))
	err = service.Listen()
	if err != nil {
		fmt.Println("Error during listen:", err)
	}
}
