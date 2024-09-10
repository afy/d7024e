package kademlia

// ===== REQ UDP PROTOCOL
// Byte 0:
//		0x0: PING
//		0x1: STORE
//		0x2: FINDNODE
//		0x3: FINDVAL
// byte 1-2:
// 		Validity UUID (byte 1: random, byte 2: port)
// Byte 3:
//		Param 1 Length (In bytes)
// Byte 4:
//		Param 2 Length (In bytes)
// Bytes 5-END:
//		Param data stream
//
// Max message size = 256*2 + 3 bytes

// ===== RESP UDP PROTOCOL
// Byte 1-END: Data
// (Per-request data is not stored in-flight)

import (
	"fmt"
	"log"
	"net"
)

const MAX_REQ_BYTES int = 512 + 5
const RESP_PORTS_RANGE_MIN = 10_000
const RESP_PORTS_RANGE_MAX = 10_100

const (
	RPC_PING     byte = 0x0
	RPC_STORE    byte = 0x1
	RPC_FINDNODE byte = 0x2
	RPC_FINDVAL  byte = 0x3
)

// Object containing all information needed for
// inter-node communication
type Network struct {
	routing_table *RoutingTable
	ports_status  []bool
}

func ParseData(buf []byte, n int) (byte, []byte, [][]byte) {
	// data := strings.TrimSpace(string(buf[:n]))
	var params = [][]byte{
		buf[3:n],
		[]byte("0"),
	}
	return buf[0], buf[1:3], params
}

func NewNetwork(ip string, port string) *Network {
	return &Network{
		NewRoutingTable(
			NewContact(
				NewRandomKademliaID(),
				ip+":"+port,
			),
		), []bool{},
	}
}

func (network *Network) Listen() *Network {
	conn, err := net.ListenPacket("udp", network.routing_table.me.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
			continue
		}
		rpc, _, _ := ParseData(buf, n)
		fmt.Printf("received: %x from %s\n", rpc, addr)

		switch rpc {

		case RPC_PING:
			fmt.Println("ping")

		case RPC_STORE:
			fmt.Println("store")

		case RPC_FINDNODE:
			fmt.Print("findnode")

		case RPC_FINDVAL:
			fmt.Println("findval")

		default:
			log.Fatal("Invalid RPC")
			continue
		}
	}
}

func (network *Network) JoinNetwork(init_addr string) {
	conn, err := net.Dial("udp", init_addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var p = []byte{
		RPC_FINDNODE, // 2 = node lookup
		0xff,         // UUID random component
		0xff,         // UUID iter component
	}
	var pbody []byte
	pbody = []byte(init_addr)
	packet := append(p, pbody...)

	_, err = conn.Write(packet)
	if err != nil {
		log.Fatal(err)
	}
}

func (network *Network) ListenReply(ip string, expected_rpc int, uuid [2]byte) {
	// TODO
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
