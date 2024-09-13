package kademlia

import (
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
	ports := make([]*PortData, PORTS_RANGE_MAX-PORTS_RANGE_MIN+1)
	for pi, p := range ports {
		p.iter = 0x00
		p.open = false
		p.num = PORTS_RANGE_MIN + pi
		p.num_str = strconv.Itoa(p.num)
	}
	return &Network{rtable, ports}
}

// Send a request to the bootstrap node (init_addr) to join the network.
func (network *Network) JoinNetwork(init_addr string) {
	p1 := []byte(init_addr) // param 1: node address to find
	p2 := []byte{}          // param 2: none
	network.SendAndWait(init_addr, RPC_FINDNODE, p1, p2)
}

// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) SendPingMessage(contact *Contact) {
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) SendFindContactMessage(contact *Contact) {
}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) SendFindDataMessage(hash string) {
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) SendStoreMessage(data []byte) {
}
