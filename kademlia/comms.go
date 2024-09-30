package kademlia

import (
	"encoding/json"
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
const PARAM_K = 20           // "k" value specified in original paper
const (
	RPC_NIL         byte = 0x0
	RPC_PING        byte = 0x2
	RPC_STORE       byte = 0x3
	RPC_FINDCONTACT byte = 0x4
	RPC_FINDVAL     byte = 0x5
)

// Contains port information, such as auth iter count.
type PortData struct {
	num     int
	num_str string
	open    bool
}

// Object containing all information needed for inter-node communication.
type Network struct {
	routing_table *RoutingTable
	dynamic_ports []*PortData
	data_store    *Store
}

type NetworkMessage struct {
	Rpc         byte     `json:"rpc"`
	Src_node_id string   `json:"src_node_id"`
	Aid         string   `json:"aid"`
	Data        [][]byte `json:"data"`
}

// Wrapper func for json data sent over network
func NewNetworkMessage(rpc byte, node_id *KademliaID, auth_id *AuthID, data [][]byte) *NetworkMessage {
	return &NetworkMessage{rpc, node_id.String(), auth_id.String(), data}
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
			true,
		}
	}
	store := NewStore()
	return &Network{rtable, ports[:], store}
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
func ParseInput(buf []byte, n int) *NetworkMessage {
	fmt.Printf("%s\n", string(buf))
	var m NetworkMessage
	err := json.Unmarshal(buf[:n], &m)
	AssertAndCrash(err)
	return &m
}

// Send a UDP packet to a node/client. Then, start waiting for a UDP packet on same port.
func (network *Network) SendAndWait(dist_ip string, rpc byte, params [][]byte) [][]byte {
	chan_msg := make(chan [][]byte)

	go func() {
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
		aid_req := GenerateAuthID()
		msg := NewNetworkMessage(rpc, network.routing_table.me.ID, aid_req, params)
		msg_bytes, err := json.Marshal(msg)
		fmt.Printf("%s\n", string(msg_bytes))
		AssertAndCrash(err)
		_, err = req_conn.Write(msg_bytes)
		req_conn.Close()
		AssertAndCrash(err)

		// Wait for response, where the auth id:s match
		resp_conn, err := net.ListenPacket("udp", ":"+req_port.num_str)
		AssertAndCrash(err)
		defer resp_conn.Close()
		fmt.Printf("RPC Listener: Waiting on %s\n", ":"+req_port.num_str)

		ret_buf := make([][]byte, MAX_PACKET_SIZE)
		for {
			resp_buf := make([]byte, MAX_PACKET_SIZE)
			n, _, err := resp_conn.ReadFrom(resp_buf)
			AssertAndCrash(err)
			var ret_msg *NetworkMessage
			errd := json.Unmarshal(resp_buf[:n], &ret_msg)
			AssertAndCrash(errd)

			if ret_msg.Aid == aid_req.String() {
				fmt.Println("RPC Listener: Response recieved")
				ret_buf = ret_msg.Data[:]
				break
			}
		}

		req_port.open = true
		chan_msg <- ret_buf
	}()

	return <-chan_msg
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
func (network *Network) SendResponse(aid *AuthID, dist_ip string, response [][]byte) {
	msg := NewNetworkMessage(RPC_NIL, network.routing_table.me.ID, aid, response)
	msg_bytes, err := json.Marshal(msg)
	AssertAndCrash(err)
	network.Send(dist_ip, msg_bytes)
}

// network.Send but with RPC parsing
// Essentially SendAndWait without response handling
func (network *Network) SendRPC(dist_ip string, rpc byte, params [][]byte) {
	aid_req := GenerateAuthID()
	msg := NewNetworkMessage(rpc, network.routing_table.me.ID, aid_req, params)
	msg_bytes, err := json.Marshal(msg)
	AssertAndCrash(err)
	network.Send(dist_ip, msg_bytes)
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
		msg := ParseInput(buf, n)
		var aid_bytes [20]byte
		copy([]byte(msg.Aid)[:], aid_bytes[:20])
		aid := NewAuthID(aid_bytes)
		fmt.Printf("Main listener: Received: %s (%x) from %s\n", GetRPCName(msg.Rpc), msg.Rpc, addr)

		// Update routing table
		network.routing_table.AddContact(NewContact(NewKademliaID(msg.Src_node_id), addr.String()))
		fmt.Printf("%+v\n", msg)

		switch msg.Rpc {

		case RPC_PING:
			target := strings.TrimSpace(string(msg.Data[0]))
			network.ManagePingMessage(aid, addr.String(), target)

		case RPC_STORE:
			network.ManageStoreMessage(aid, addr.String(), string(msg.Data[0]), string(msg.Data[1]))

		case RPC_FINDCONTACT:
			target := strings.TrimSpace(string(msg.Data[0]))
			network.ManageFindContactMessage(aid, addr.String(), target)

		case RPC_FINDVAL:
			target := strings.TrimSpace(string(msg.Data[0]))
			network.ManageFindDataMessage(aid, addr.String(), target)

		default:
			fmt.Printf("Main listener: Invalid RPC: %s\n", string(msg.Rpc))
		}
	}
}
