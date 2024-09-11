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
	"strings"
)

const MAX_REQ_BYTES int = 512 + 5
const PORTS_RANGE_MIN = 10_000
const PORTS_RANGE_MAX = 10_100

const (
	RPC_PING     byte = 0x0
	RPC_STORE    byte = 0x1
	RPC_FINDNODE byte = 0x2
	RPC_FINDVAL  byte = 0x3
)

// Object containing all information needed for inter-node communication.
// ports_iter contains the iteration byte of every port in the usable range.
type Network struct {
	routing_table *RoutingTable
	ports_status  []byte
}

// Parse *incoming* data according to protocol at top of file.
func ParseInput(buf []byte, n int) (byte, AuthUUID, []string) {
	var (
		rpc_code byte = buf[0]
		uid_0    byte = buf[1]
		uid_1    byte = buf[2]
		p1_len   byte = buf[3]
	)
	param_1 := strings.TrimSpace(string(buf[5 : 5+p1_len]))
	param_2 := strings.TrimSpace(string(buf[5+p1_len+1 : n+1])) // note: p2 not technically needed here; review how to document this
	return rpc_code, NewAuthUUID(uid_0, uid_1), []string{param_1, param_2}
}

// Create a new Network instance with random id.
func NewNetwork(this_ip string, port string) *Network {
	rtable := NewRoutingTable(NewContact(NewRandomKademliaID(), this_ip+":"+port))
	ports := make([]byte, PORTS_RANGE_MAX-PORTS_RANGE_MIN)
	return &Network{rtable, ports}
}

// Primary listening loop at UDP, default port in [project root]/.env.
// Listen for incoming requests and handle accordingly.
// NOTE: responses go to a different port.
func (network *Network) Listen() *Network {
	conn, err := net.ListenPacket("udp", network.routing_table.me.Address)
	AssertAndCrash(err)
	defer conn.Close()
	fmt.Printf("Listening for requests on %s\n", network.routing_table.me.Address)

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
			continue
		}
		rpc, _, _ := ParseInput(buf, n)
		fmt.Printf("received: %x from %s\n", rpc, addr)

		switch rpc {

		case RPC_PING:
			fmt.Println("ping")

		case RPC_STORE:
			fmt.Println("store")

		case RPC_FINDNODE:
			fmt.Println("findnode")

		case RPC_FINDVAL:
			fmt.Println("findval")

		default:
			log.Fatal("Invalid RPC: " + string(rpc))
		}
	}
}

// Send a UDP packet to a node/client. Then, start waiting for a UDP packet on same port.
// Response will be returned as a byte array.
func (network *Network) SendAndWait(dist_ip string, rpc byte, param_1 []byte, param_2 []byte) []byte {
	conn, err := net.Dial("udp", dist_ip)
	AssertAndCrash(err)
	defer conn.Close()
	fmt.Printf("Sent RPC: %x to %s\n", rpc, dist_ip)

	auid := GenerateAuthUUID(0xff)

	// port := ParsePortNumber(conn.LocalAddr().String())
	// fmt.Printf("%d\n", port)

	// Format network packet (see docs)
	var packet = []byte{
		rpc,           // 2 = node lookup
		auid.value[0], // UUID random component
		auid.value[1], // UUID iter component
	}
	len_p1 := len(string(param_1))
	len_p2 := len(string(param_2))
	body := []byte{
		byte(len_p1),
		byte(len_p2),
	}
	body = append(body, param_1...)
	body = append(body, param_2...)
	packet = append(packet, body...)
	_, err = conn.Write(packet)
	AssertAndCrash(err)

	// Wait for response
  resp_port := 10000
  resp_addr := fmt.Sprintf(":%d", resp_port)
	resp_conn, err := net.ListenPacket("udp", resp_addr)
	AssertAndCrash(err)
	defer resp_conn.Close()
	fmt.Printf("Listening for response on %s\n", resp_addr)

  resp_buf := make([]byte, 1024)
  n, addr, err := resp_conn.ReadFrom(resp_buf)
  AssertAndCrash(err)
  resp_data := strings.TrimSpace(string(resp_buf[:n]))
  fmt.Printf("received response: %s from %s\n", resp_data, addr)

	// return data
  return resp_buf[:n]
}

// Send a request to the bootstrap node (init_addr) to join the network.
func (network *Network) JoinNetwork(init_addr string) {
	p1 := []byte(init_addr) // param 1: node address to find
	p2 := []byte{}          // param 2: none
	network.SendAndWait(init_addr, RPC_FINDNODE, p1, p2)
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
