package kademlia

import (
	"fmt"
	"net"
	"strconv"
)

// Object containing all information needed for inter-node communication.
// ports_iter contains the iteration byte of every port in the usable range.
type Network struct {
	routing_table *RoutingTable
	dynamic_ports []*PortData
}

// Create a new Network instance with random id.
func NewNetwork(this_ip string, port string) *Network {
	addr := this_ip + ":" + port
	rtable := NewRoutingTable(NewContact(NewRandomKademliaID(), addr))
	var ports [PRANGE_MAX - PRANGE_MIN + 1]*PortData
	for pi, _ := range ports {
		ports[pi] = &PortData{
			PRANGE_MIN + pi,
			strconv.Itoa(PRANGE_MIN + pi),
			0x00,
			false,
		}
	}
	return &Network{rtable, ports[:]}
}

// Send a request to the bootstrap node (init_addr) to join the network.
func (network *Network) JoinNetwork(init_addr string) {
	p1 := []byte(init_addr) // param 1: node address to find
	p2 := []byte{}          // param 2: none
	resp := network.SendAndWait(init_addr, RPC_FINDNODE, p1, p2)
	fmt.Printf("Response from server: %s\n", string(resp))
}

// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) SendPingMessage(meta *MessageMetadata, target_id []byte) {
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) SendFindContactMessage(meta *MessageMetadata, target_id []byte) {
	c, err := net.Dial("udp", meta.addr)
	AssertAndCrash(err)
	defer c.Close()
	fmt.Printf("Sending response to %s\n", meta.addr)
	c.Write([]byte("RESPONSE HERE"))
}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) SendFindDataMessage(meta *MessageMetadata, target_id []byte) {

}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) SendStoreMessage(meta *MessageMetadata, target_id []byte, value []byte) {
}
