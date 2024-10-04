package kademlia

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const MAX_PACKET_SIZE = 1024 // UDP packet buffer size.
const PRANGE_MIN = 10_000    // Lower component of port range.
const PRANGE_MAX = 10_100    // Upper component of port range.
const ALPHA = 3              // For node lookup; how many nodes to query
const (
	RPC_PING        byte = 0x0
	RPC_STORE       byte = 0x1
	RPC_FINDCONTACT byte = 0x2
	RPC_FINDVAL     byte = 0x3
)

// Contains port information, such as auth iter count.
type PortData struct {
	num     int
	num_str string
	iter    byte
	open    bool
}

// Object containing all information needed for inter-node communication.
type Network struct {
	routing_table *RoutingTable
	dynamic_ports []*PortData
}

func (network *Network) GetID() string {
  return network.routing_table.me.ID.String()
}

// Create a new Network instance with random id,
// Unless it is the bootstrap node, whose nodeid is configured in the .env file.
func NewNetwork(this_ip string, port string) *Network {
	addr := this_ip + ":" + port
	is_bootstrap, _ := strconv.ParseBool(os.Getenv("IS_BOOTSTRAP_NODE"))
	var rtable *RoutingTable
	if !is_bootstrap {
		rtable = NewRoutingTable(NewContact(NewRandomKademliaID(), addr))
	} else {
		rtable = NewRoutingTable(NewContact(NewKademliaID(os.Getenv("BOOTSTRAP_NODE_ID")), addr))
	}
	var ports [PRANGE_MAX - PRANGE_MIN + 1]*PortData
	for pi := range ports {
		ports[pi] = &PortData{
			PRANGE_MIN + pi,
			strconv.Itoa(PRANGE_MIN + pi),
			0x00,
			true,
		}
	}
	return &Network{rtable, ports[:]}
}

// Get the first open port from the dynamic_ports list.
func (network *Network) GetFirstOpenPort() *PortData {
	max_ind := PRANGE_MAX - PRANGE_MIN
	for i := 0; i <= max_ind; i++ {
		port := network.dynamic_ports[i]

		// Using port.open is more reliable than attempting to bind.
		// It is read-only here, port should only be set in SendAndWait()
		if port.open {
			return port
		}
	}
	panic("No open ports!")
}

// Parse *incoming* request data in main listener according to protocol at top of file.
func ParseInput(buf []byte, n int) (byte, *AuthID, [][]byte) {
	var (
		rpc_code byte = buf[0]
		uid_0    byte = buf[1]
		uid_1    byte = buf[2]
		p1_len   byte = buf[3]
	)
	param_1 := buf[5 : 5+p1_len]
	param_2 := buf[5+p1_len+1 : n+1] // note: p2 not technically needed here; review how to document this
	auth := NewAuthID(uid_0, uid_1)
	return rpc_code, &auth, [][]byte{param_1, param_2}
}

// Send a UDP packet to a node/client. Then, start waiting for a UDP packet on same port.
func (network *Network) SendAndWait(dist_ip string, rpc byte, param_1 []byte, param_2 []byte) []byte {
	req_port := network.GetFirstOpenPort()
	req_port.open = false

	addr, err := net.ResolveUDPAddr("udp", ":"+req_port.num_str)
	AssertAndCrash(err)
	dialer := net.Dialer{
		LocalAddr: addr,
		Timeout:   time.Duration(5 * float64(time.Second)), // great design choice
	}

	// No defer; close connection directly after sending UDP packet
	req_conn, err := dialer.Dial("udp", dist_ip)
	AssertAndCrash(err)
	fmt.Printf("RPC Listener: Sent RPC %s to %s from %s\n", GetRPCName(rpc), dist_ip, ":"+req_port.num_str)

	// Format network packet (see docs)
	aid_req := GenerateAuthID(req_port.iter)
	req_port.iter += 0x01
	len_p1 := len(string(param_1))
	len_p2 := len(string(param_2))
	body := []byte{
		rpc,              // 2 = node lookup
		aid_req.value[0], // UUID random component
		aid_req.value[1], // UUID iter component
		byte(len_p1),     // #bytes first arg
		byte(len_p2),     // #bytes second arg
	}
	body = append(body, param_1...)
	body = append(body, param_2...)
	_, err = req_conn.Write(body)
	req_conn.Close()
	AssertAndCrash(err)

	// Wait for response, where the auth id:s match
	resp_conn, err := net.ListenPacket("udp", ":"+req_port.num_str)
	AssertAndCrash(err)
	defer resp_conn.Close()
	fmt.Printf("RPC Listener: Waiting on %s\n", ":"+req_port.num_str)

	ret_buf := make([]byte, MAX_PACKET_SIZE)
	for {
		resp_buf := make([]byte, MAX_PACKET_SIZE)
		_, _, err := resp_conn.ReadFrom(resp_buf)
		AssertAndCrash(err)

		aid_resp := NewAuthID(resp_buf[0], resp_buf[1])
		if aid_resp.Equals(aid_req) {
			fmt.Println("RPC Listener: Response recieved")
			ret_buf = resp_buf[2:]
			break
		}
	}

	req_port.open = true
	return ret_buf
}

// Send function to send a response back to the specified address.
// Never use in implementation, rather use SendResponse or SendRPC
func (network *Network) Send(dist_ip string, response []byte) {
	resp_addr, err := net.ResolveUDPAddr("udp", dist_ip)
	if err != nil {
		fmt.Printf("Error resolving %s: %v\n", dist_ip, err)
		return
	}
	conn, err := net.DialUDP("udp", nil, resp_addr)
	if err != nil {
		fmt.Printf("Error dialing UDP: %vn", err)
		return
	}
	defer conn.Close()
	_, err = conn.Write(response)
	if err != nil {
		fmt.Printf("Error sending response: %vn", err)
	} else {
		fmt.Printf("Response sent to: %v\n", dist_ip)
	}
}

// network.Send but with AID for responses
func (network *Network) SendResponse(aid *AuthID, dist_ip string, response []byte) {
	data := append(aid.value[:], response...)
	network.Send(dist_ip, data)
}

// network.Send but with RPC parsing
// Essentially SendAndWait without response handling
func (network *Network) SendRPC(dist_ip string, rpc byte, param_1 []byte, param_2 []byte) {
	aid_req := GenerateAuthID(0x00)
	len_p1 := len(string(param_1))
	len_p2 := len(string(param_2))
	body := []byte{
		rpc,              // 2 = node lookup
		aid_req.value[0], // UUID random component
		aid_req.value[1], // UUID iter component
		byte(len_p1),     // #bytes first arg
		byte(len_p2),     // #bytes second arg
	}
	body = append(body, param_1...)
	body = append(body, param_2...)
	network.Send(dist_ip, body)
}

// Primary listening loop at UDP, default port in [project root]/.env.
// Listen for incoming requests and handle accordingly.
// NOTE: responses go to a different port.
func (network *Network) Listen() *Network {
	conn, err := net.ListenPacket("udp", network.routing_table.me.Address)
	AssertAndCrash(err)
	defer conn.Close()
	fmt.Printf("Main listener: Listening for requests on %s\n", network.routing_table.me.Address)

	for {
		buf := make([]byte, MAX_PACKET_SIZE)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
			continue
		}
		rpc, aid, params := ParseInput(buf, n)
		fmt.Printf("Main listener: Received: %s from %s\n", GetRPCName(rpc), addr)

		switch rpc {

		case RPC_PING:
			target := strings.TrimSpace(string(params[0]))
			network.ManagePingMessage(aid, addr.String(), target)

		case RPC_STORE:
			network.ManageStoreMessage(aid, addr.String(), params[0], params[1])

		case RPC_FINDCONTACT:
			target := strings.TrimSpace(string(params[0]))
			network.ManageFindContactMessage(aid, addr.String(), target)

		case RPC_FINDVAL:
			target := strings.TrimSpace(string(params[0]))
			network.ManageFindDataMessage(aid, addr.String(), target)

		default:
			fmt.Printf("Main listener: Invalid RPC: %s\n", string(rpc))
		}
	}
}
