package bootstrap

import (
	"io"

	"github.com/golemfactory/bootstrap_go/cbor"
	cborimpl "github.com/whyrusleeping/cbor/go"
)

type Node struct {
	NodeName     string        `pyobj:"node_name"`
	Key          string        `pyobj:"key"`
	PrvPort      uint64        `pyobj:"prv_port"`
	PubPort      uint64        `pyobj:"pub_port"`
	P2pPrvPort   uint64        `pyobj:"p2p_prv_port"`
	P2pPubPort   uint64        `pyobj:"p2p_pub_port"`
	PrvAddr      string        `pyobj:"prv_addr"`
	PubAddr      string        `pyobj:"pub_addr"`
	PrvAddresses []interface{} `pyobj:"prv_addresses"`
	NatType      string        `pyobj:"nat_type"`
}

func (self *Node) GetPyObjectName() string {
	return "golem.network.p2p.node.Node"
}

func (self *Node) ToCBOR(w io.Writer, enc *cborimpl.Encoder) error {
	return cbor.ToCBOR(self, w, enc)
}

func init() {
	cbor.RegisterPythonType("golem.network.p2p.node.Node", func() interface{} { return &Node{} })
}
