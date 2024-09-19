package kademlia

import (
	"net"
	"strconv"
	"fmt"
)

// Object containing all information needed for inter-node communication.
type Network struct {
	routing_table *RoutingTable
	dynamic_ports []*PortData
}

// Create a new Network instance with random id.
func NewNetwork(this_ip string, port string) *Network {
	addr := this_ip + ":" + port
	rtable := NewRoutingTable(NewContact(NewRandomKademliaID(), addr))
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

	// %%%%%%%%% TEST FUNCTION, rewrite this
	c, err := net.Dial("udp", meta.addr)
	AssertAndCrash(err)
	defer c.Close()
	fmt.Printf("Sending response to %s\n", meta.addr)

	var body []byte
	body = append([]byte{}, meta.auuid.value[:]...)
	body = append(body, []byte("RESPONSE HERE")...)
	c.Write(body)
}

const alpha = 8

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) SendFindDataMessage(meta *MessageMetadata, target_id []byte) ([]byte, error) {

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
			return response[1:], nil // Return the found data, skipping the success byte
		} else if len(response) > 1 && response[0] == 0x00 {
				// Target is a node, not data
				fmt.Printf("Target %x is a node; continuing search.\n", target_id)
				continue
		}
	}

	// If no data or node is found in any response
	return nil, fmt.Errorf("Data not found for target %x", targetID)
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) SendStoreMessage(meta *MessageMetadata, target_id []byte, value []byte) {
}
