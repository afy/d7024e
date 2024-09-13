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
	num     int
	num_str string
	iter    byte
	open    bool
}

// Parse from Listen(), send to RPC handlers to simplify general func params.
type MessageMetadata struct {
	auuid *AuthUUID
	addr  string
}

// Return Message metadata instance for main listener.
func NewMessageMetadata(uuid *AuthUUID, addr string) MessageMetadata {
	return MessageMetadata{uuid, addr}
}

// Get the first open port from the dynamic_ports list.
func (network *Network) GetFirstOpenPort() *PortData {
	max_ind := PRANGE_MAX - PRANGE_MIN
	for i := 0; i <= max_ind; i++ {
		port := network.dynamic_ports[i]

		// Using port.open is more reliable than attempting to bind.
		// It is read-only here, port should only be set in SendAndWait()
		if !port.open {
			return port
		}
	}
	panic("No open ports!")
}

// Parse *incoming* data in main listener according to protocol at top of file.
func ParseInput(buf []byte, n int) (byte, *AuthUUID, [][]byte) {
	var (
		rpc_code byte = buf[0]
		uid_0    byte = buf[1]
		uid_1    byte = buf[2]
		p1_len   byte = buf[3]
	)
	param_1 := buf[5 : 5+p1_len]
	param_2 := buf[5+p1_len+1 : n+1] // note: p2 not technically needed here; review how to document this
	auth := NewAuthUUID(uid_0, uid_1)
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
	req_port.open = false
	defer func(port *PortData) { port.open = true }(req_port)

	addr, err := net.ResolveUDPAddr("udp", ":"+req_port.num_str)
	AssertAndCrash(err)
	dialer := net.Dialer{
		LocalAddr: addr,
	}

	// No defer; close connection directly after sending UDP packet
	req_conn, err := dialer.Dial("udp", dist_ip)
	AssertAndCrash(err)
	fmt.Printf("RPC Listener: Sent RPC %x to %s from %s\n", rpc, dist_ip, ":"+req_port.num_str)

	// Format network packet (see docs)
	auid_req := GenerateAuthUUID(req_port.iter)
	req_port.iter += 0x01
	len_p1 := len(string(param_1))
	len_p2 := len(string(param_2))
	body := []byte{
		rpc,               // 2 = node lookup
		auid_req.value[0], // UUID random component
		auid_req.value[1], // UUID iter component
		byte(len_p1),      // #bytes first arg
		byte(len_p2),      // #bytes second arg
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
		n, addr, err := resp_conn.ReadFrom(resp_buf)
		AssertAndCrash(err)
		resp_data := strings.TrimSpace(string(resp_buf[:n]))
		fmt.Printf("RPC Listener: Received response: %s from %s\n", resp_data, addr)

		auid_resp := NewAuthUUID(resp_buf[0], resp_buf[1])
		if auid_resp.Equals(auid_req) {
			fmt.Println("RPC Listener: Matching AAUID; breaking loop")
			ret_buf = resp_buf[2:]
			break
		}
	}

	return ret_buf
}
