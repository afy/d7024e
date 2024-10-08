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
// const PRANGE_MIN = 10_000    // Lower component of port range.
const MAX_PORTS = 100    // Upper component of port range.
const ALPHA = 3              // For node lookup; how many nodes to query
const PARAM_K = 20           // "k" value specified in original paper
const (
	// RPC Codes (byte[0] = 0)
	RPC_NIL         byte = 0x00
	RPC_PING        byte = 0x01
	RPC_STORE       byte = 0x02
	RPC_FINDCONTACT byte = 0x03
	RPC_FINDVAL     byte = 0x04
	RPC_NODELOOKUP  byte = 0x05

	// RPC Response codes (byte[0] = F)
	RESP_VALFOUND     byte = 0xF0 // From store/findval, indicating value returned
	RESP_CONTACTS     byte = 0xF1 // From findval/contact indicating a list of contacts
	RESP_STORE_OK     byte = 0xF2 // Store has been a sucess
	RESP_STORE_EXISTS byte = 0xF3 // Value already exists in the network
	RESP_PING_OK      byte = 0xF4 // PING response
	RESP_PING_FAIL    byte = 0xF5
)

type byte_arr_list [][]byte

// Object containing all information needed for inter-node communication.
type Network struct {
	routing_table *RoutingTable
	data_store    *Store
  min_port      int
  offset_port   int
}

type NetworkMessage struct {
	Rpc         byte          `json:"rpc"`
	Src_node_id string        `json:"src_node_id"`
  Src_port    int           `json:"src_port"`
	Aid         string        `json:"aid"`
	Data        byte_arr_list `json:"data"`
}

// Wrapper func for json data sent over network
func NewNetworkMessage(rpc byte, node_id *KademliaID, src_port int, auth_id *AuthID, data byte_arr_list) *NetworkMessage {
	return &NetworkMessage{rpc, node_id.String(), src_port, auth_id.String(), data}
}

func (network *Network) GetID() string {
	return network.routing_table.me.ID.String()
}

func (network *Network) GetPort() int {
  _, port := ParsePortNumber(network.routing_table.me.Address)
  return port
}

// Create a new Network instance with random id,
// Unless it is the bootstrap node, whose nodeid is configured in the .env file.
func NewNetwork(this_ip string, port string, min_port int) *Network {
	addr := this_ip + ":" + port
	is_bootstrap, _ := strconv.ParseBool(os.Getenv("IS_BOOTSTRAP_NODE"))
	var rtable *RoutingTable
	if !is_bootstrap {
		rtable = NewRoutingTable(NewContact(NewRandomKademliaID(), addr))
	} else {
		rtable = NewRoutingTable(NewContact(NewKademliaID(os.Getenv("BOOTSTRAP_NODE_ID")), addr))
	}
	
	store := NewStore()
	fmt.Printf("NodeId: %s\n", rtable.me.ID.String())
	return &Network{rtable, store, min_port, 0}
}

func (network *Network) GetNextPort() int {
  offset := network.offset_port
  network.offset_port++
  if network.offset_port >= MAX_PORTS {
    network.offset_port = 0
  }
  return network.min_port + offset
}

// Send a UDP packet to a node/client. Then, start waiting for a UDP packet on same port.
func (network *Network) SendAndWait(dist_ip string, rpc byte, params byte_arr_list) NetworkMessage {
	chan_msg := make(chan NetworkMessage)
	go func() {
    req_port := network.GetNextPort()
    max_retries := 10
    retry_count := 0

    var req_conn net.Conn

    for retry_count < max_retries {
      addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", req_port))
		  AssertAndCrash(err)
		  dialer := net.Dialer{
			  LocalAddr: addr,
			  Timeout:   time.Duration(5 * float64(time.Second)), // great design choice
		  }
      
      req_conn, err = dialer.Dial("udp", dist_ip)

      if err != nil && err.Error() == "bind: address already in use" {
        retry_count++
        continue
      }
      
      if err == nil {
        break
      }
      
      AssertAndCrash(err)
    }

		// No defer; close connection directly after sending UDP packet
    fmt.Printf("RPC: Sent RPC %s to %s from :%d\n", GetRPCName(rpc), dist_ip, req_port)

		// Format network packet (see docs)
		aid_req := GenerateAuthID()
		msg := NewNetworkMessage(rpc, network.routing_table.me.ID, network.GetPort(), aid_req, params)
		msg_bytes, err := json.Marshal(msg)
		AssertAndCrash(err)

		_, err = req_conn.Write(msg_bytes)
		req_conn.Close()
		AssertAndCrash(err)

		// Wait for response, where the auth id:s match
		resp_conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", req_port))
		AssertAndCrash(err)
		defer resp_conn.Close()
    fmt.Printf("RPC: Waiting on :%d\n", req_port)

		for {
			resp_buf := make([]byte, MAX_PACKET_SIZE)
			n, _, err := resp_conn.ReadFrom(resp_buf)
			AssertAndCrash(err)

			var ret_msg *NetworkMessage
			errd := json.Unmarshal(resp_buf[:n], &ret_msg)
			AssertAndCrash(errd)

			if ret_msg.Aid == aid_req.String() {
        fmt.Printf("RPC: Response recieved on :%d\n", req_port)
				chan_msg <- *ret_msg
				break
			}
		}
	}()
	return <-chan_msg
}

// Send function to send a response back to the specified address.
// Never use in implementation, rather use SendResponse or SendRPC
func (network *Network) Send(dist_ip string, response *NetworkMessage) {
	resp_addr, err := net.ResolveUDPAddr("udp", dist_ip)
	resp_bytes, err := json.Marshal(response)
	AssertAndCrash(err)
	if err != nil {
		fmt.Printf("RPC: Error resolving %s: %v\n", dist_ip, err)
		return
	}
	conn, err := net.DialUDP("udp", nil, resp_addr)
	if err != nil {
		fmt.Printf("RPC: Error dialing UDP: %vn", err)
		return
	}
	defer conn.Close()
	_, err = conn.Write(resp_bytes)
	if err != nil {
		fmt.Printf("RPC: Error sending response: %vn", err)
	} else {
		fmt.Printf("RPC: Response sent to: %v\n", dist_ip)
	}
}

// network.Send but with AID for responses
func (network *Network) SendResponse(aid *AuthID, dist_ip string, response_rpc byte, response []byte) {
	if response_rpc&0xF0 != 0xF0 {
		fmt.Println("Warning: response rpc in SendResponse: does not seem to be of type response (see comms.go)")
	}
	resp := make(byte_arr_list, 1)
	resp[0] = response
  msg := NewNetworkMessage(response_rpc, network.routing_table.me.ID, network.GetPort(), aid, resp)
	network.Send(dist_ip, msg)
}

// network.Send but with RPC parsing
// Essentially SendAndWait without response handling
func (network *Network) SendRPC(dist_ip string, rpc byte, params byte_arr_list) {
	aid_req := GenerateAuthID()
	msg := NewNetworkMessage(rpc, network.routing_table.me.ID, network.GetPort(), aid_req, params)
	network.Send(dist_ip, msg)
}

// Primary listening loop at UDP, default port in [project root]/.env.
// Listen for incoming requests and handle accordingly.
func (network *Network) Listen() *Network {
	conn, err := net.ListenPacket("udp", network.routing_table.me.Address)
	AssertAndCrash(err)
	defer conn.Close()
	fmt.Printf("Main: Listening for requests on %s\n", network.routing_table.me.Address)

	for {
		buf := make([]byte, MAX_PACKET_SIZE)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}
    // TODO: Move to separate function
		var msg NetworkMessage
		err2 := json.Unmarshal(buf[:n], &msg)
		if err2 != nil {
			log.Println(err)
			continue
		}

		var aid_bytes [20]byte
		copy([]byte(msg.Aid)[:], aid_bytes[:20])
		aid := NewAuthID(aid_bytes)
		fmt.Printf("Main: Received: %s (%x) from %s (%s)\n", GetRPCName(msg.Rpc), msg.Rpc, msg.Src_node_id, addr)

		// Update routing table
		src_ip, _ := ParsePortNumber(addr.String())
    network.routing_table.AddContact(NewContact(NewKademliaID(msg.Src_node_id), fmt.Sprintf("%s:%d", src_ip, msg.Src_port)))
		for _, b := range network.routing_table.buckets {
			for e := b.list.Front(); e != nil; e = e.Next() {
				fmt.Printf("%s\n", e.Value)
			}
		}

		switch msg.Rpc {
		case RPC_PING:
			target := strings.TrimSpace(string(msg.Data[0]))
			go network.ManagePing(aid, addr.String(), target)

		case RPC_STORE:
			go network.ManageStore(aid, addr.String(), string(msg.Data[0]), string(msg.Data[1]))

		case RPC_FINDCONTACT:
			target := strings.TrimSpace(string(msg.Data[0]))
			go network.ManageFindContact(aid, addr.String(), target)

		case RPC_FINDVAL:
			target := strings.TrimSpace(string(msg.Data[0]))
			go network.ManageFindData(aid, addr.String(), target)

		case RPC_NODELOOKUP:
			target := strings.TrimSpace(string(msg.Data[0]))
			go network.ManageNodeLookup(aid, addr.String(), target)

		default:
			fmt.Printf("Main: Invalid RPC: %s\n", string(msg.Rpc))
		}
	}
}
