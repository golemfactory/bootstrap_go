package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	bootstrap "github.com/golemfactory/bootstrap_go"
	"github.com/golemfactory/bootstrap_go/crypto"
	"github.com/golemfactory/bootstrap_go/peerkeeper"

	"github.com/ccding/go-stun/stun"
)

const (
	PORT                   = 40102
	PEER_NUM               = 100
	NAME                   = "Go Bootstrap"
	PROTO_ID               = "31"
	GOLEM_MESSAGES_VERSION = "2.24.3"
	GOLEM_VERSION          = "0.19.0"
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

	_, host, err := stun.NewClient().Discover()
	if err != nil {
		fmt.Println("Error discovering STUN details:", err)
		return
	}

	privKey, err := crypto.GenerateDifficultKey(KEY_DIFF)
	if err != nil {
		log.Println("Error while generating private key", err)
		return
	}
	pubKey := privKey.GetPublicKey()

	config := &bootstrap.Config{
		Name:                 name,
		Id:                   pubKey.Hex(),
		Port:                 port,
		PrvAddr:              prvAddresses[0].(string),
		PubAddr:              host.IP(),
		PrvAddresses:         prvAddresses,
		NatType:              make([]interface{}, 0),
		PeerNum:              PEER_NUM,
		KeyDifficulty:        KEY_DIFF,
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
