package kademlia

// File used to simplify network.go
// by collecting all functionality that support the RPCs,
// like UDP listeners and support structs.

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
// Byte 0-1: Auth UUID
// byte 2-END: Data

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const MAX_PACKET_SIZE = 1024 // UDP packet buffer size.
const PRANGE_MIN = 10_000    // Lower component of port range.
const PRANGE_MAX = 10_100    // Upper component of port range.
const (
	RPC_PING     byte = 0x0
	RPC_STORE    byte = 0x1
	RPC_FINDNODE byte = 0x2
	RPC_FINDVAL  byte = 0x3
)

// Contains port information, such as auth iter count.
type PortData struct {
	Num     int
	Num_str string
	Iter    byte
	Open    bool
}

// Parse from Listen(), send to RPC handlers to simplify general func params.
type RequestMetadata struct {
	Id   *AuthID
	Addr string
}

// Return Message metadata instance for main listener.
func NewMessageMetadata(uuid *AuthID, addr string) RequestMetadata {
	return RequestMetadata{uuid, addr}
}

// Object containing all information needed for inter-node communication.
type Network struct {
	routing_table *RoutingTable
	dynamic_ports []*PortData
}

// Create a new Network instance with random id.
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
			false,
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
		if !port.Open {
			return port
		}
	}
	panic("No open ports!")
}

// Parse *incoming* data in main listener according to protocol at top of file.
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

// Primary listening loop at UDP, default port in [project root]/.env.
// Listen for incoming requests and handle accordingly.
//
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
		rpc, auuid, params := ParseInput(buf, n)
		fmt.Printf("Main listener: Received: %x from %s\n", rpc, addr)

		meta := NewMessageMetadata(auuid, addr.String())
		switch rpc {

		case RPC_PING:
			network.SendPingMessage(&meta, params[0])

		case RPC_STORE:
			network.SendStoreMessage(&meta, params[0], params[1])

		case RPC_FINDNODE:
			network.SendFindContactMessage(&meta, params[0])

		case RPC_FINDVAL:
			network.SendFindDataMessage(&meta, params[0])

		default:
			fmt.Printf("Main listener: Invalid RPC: %s\n", string(rpc))
		}
	}
}

// Send a UDP packet to a node/client. Then, start waiting for a UDP packet on same port.
func (network *Network) SendAndWait(dist_ip string, rpc byte, param_1 []byte, param_2 []byte) []byte {
	req_port := network.GetFirstOpenPort()
	req_port.Open = false
	defer func(port *PortData) { port.Open = true }(req_port)

	addr, err := net.ResolveUDPAddr("udp", ":"+req_port.Num_str)
	AssertAndCrash(err)
	dialer := net.Dialer{
		LocalAddr: addr,
	}

	// No defer; close connection directly after sending UDP packet
	req_conn, err := dialer.Dial("udp", dist_ip)
	AssertAndCrash(err)
	fmt.Printf("RPC Listener: Sent RPC %x to %s from %s\n", rpc, dist_ip, ":"+req_port.Num_str)

	// Format network packet (see docs)
	aid_req := GenerateAuthID(req_port.Iter)
	req_port.Iter += 0x01
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
	resp_conn, err := net.ListenPacket("udp", ":"+req_port.Num_str)
	AssertAndCrash(err)
	defer resp_conn.Close()
	fmt.Printf("RPC Listener: Waiting on %s\n", ":"+req_port.Num_str)

	ret_buf := make([]byte, MAX_PACKET_SIZE)
	for {
		resp_buf := make([]byte, MAX_PACKET_SIZE)
		n, addr, err := resp_conn.ReadFrom(resp_buf)
		AssertAndCrash(err)
		resp_data := strings.TrimSpace(string(resp_buf[:n]))
		fmt.Printf("RPC Listener: Received response: %s from %s\n", resp_data, addr)

		aid_resp := NewAuthID(resp_buf[0], resp_buf[1])
		if aid_resp.Equals(aid_req) {
			fmt.Println("RPC Listener: Matching AAUID; breaking loop")
			ret_buf = resp_buf[2:]
			break
		}
	}

	return ret_buf
}

// Send function to send a response back to the specified addres
func (network *Network) Send(addr string, response []byte) {
	responseAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Printf("Error resolving %s: %v\n", addr, err)
		return
	}

	conn, err := net.DialUDP("udp", nil, responseAddr)
	if err != nil {
		fmt.Printf("Error dialing UDP: %vn", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(response)
	if err != nil {
		fmt.Printf("Error sending response: %vn", err)
	} else {
		fmt.Printf("Response sent to: %vn", addr)
	}
}

// Send a request to the bootstrap node (init_addr) to join the network.
func (network *Network) JoinNetwork(init_addr string) {
	bootstrap_id := NewKademliaID(os.Getenv("BOOTSTRAP_NODE_ID"))
	network.routing_table.AddContact(NewContact(bootstrap_id, init_addr))
	resp := network.SendAndWait(init_addr, RPC_FINDNODE, []byte(init_addr), []byte{})
	fmt.Printf("Response from server: %s\n", string(resp))
}

// SendPingMessage handles a PING request.
// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) SendPingMessage(meta *RequestMetadata, target_id []byte) {
	target := NewKademliaID(string(target_id))
	if target.Equals(network.routing_table.me.ID) {
		fmt.Printf("Responding to PING from %s\n", meta.Addr)
		response := []byte{meta.Id.value[0], meta.Id.value[1]}
		network.Send(meta.Addr, response)
	} else {
		fmt.Printf("Target is not this node. Finding closest node to %s\n", target.String())
		closestContacts := network.routing_table.FindClosestContacts(target, 1)

		if len(closestContacts) == 0 {
			fmt.Printf("No closest node found")
			return
		}

		closestNode := closestContacts[0]
		pingMessage := append([]byte{RPC_PING}, target_id...)
		network.SendAndWait(closestNode.Address, RPC_PING, pingMessage, nil)
	}
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) SendFindContactMessage(meta *RequestMetadata, target_id []byte) {

	// %%%%%%%%% TEST FUNCTION, rewrite this
	c, err := net.Dial("udp", meta.Addr)
	AssertAndCrash(err)
	defer c.Close()
	fmt.Printf("Sending response to %s\n", meta.Addr)

	var body []byte
	body = append([]byte{}, meta.Id.value[:]...)
	body = append(body, []byte("RESPONSE HERE")...)
	c.Write(body)
}

const alpha = 8

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) SendFindDataMessage(meta *RequestMetadata, target_id []byte) {

	// Converts the bytes slice target_id to a KademliaID
	targetID := NewKademliaID(fmt.Sprintf("%x", target_id))

	// Find closest contacts using alpha
	closestContacts := network.routing_table.FindClosestContacts(targetID, alpha)

	// Iterate over closest contacts and send FindData requests
	for _, contact := range closestContacts {
		fmt.Printf("Sending FindData request to: %s\n", contact.Address)

		// Using the SendAndWait function instead of direct send
		response := network.SendAndWait(contact.Address, RPC_FINDVAL, targetID[:], []byte{})

		// Check if response is empty during the network communication failure
		if len(response) == 0 {
			fmt.Printf("No response or invalid response received from %s\n", contact.Address)
			continue // Skip to the next contact
		}

		// Check if the first byte of the response indicates data found
		if len(response) > 1 && response[0] == 0x01 {
			// Data found in the response
			fmt.Printf("Data found for target %x at %s\n", target_id, contact.Address)
			//return response[1:], nil // Return the found data, skipping the success byte
		} else if len(response) > 1 && response[0] == 0x00 {
			// Target is a node, not data
			fmt.Printf("Target %x is a node; continuing search.\n", target_id)
			continue
		}
	}

	// If no data or node is found in any response
	//return nil, fmt.Errorf("Data not found for target %x", targetID)
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) SendStoreMessage(meta *RequestMetadata, target_id []byte, value []byte) {
}
