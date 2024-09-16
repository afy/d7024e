package kademlia

import (
	"fmt"
	"net"
	"strconv"
	"time"
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

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) SendFindDataMessage(meta *MessageMetadata, target_id []byte) ([]byte, error) {

	// Converts the bytes slice target_id to a KademliaID 
	targetID := NewKademliaID(fmt.Sprintf("%x", target_id))

	// Find closest contacts using alpha 
	closestContacts := network.routing_table.FindClosestContacts(targetID, alpha)

	// Create a FindData message to be sent to the closest nodes
	findDataMessage := NewFindDataMessage(meta.auuid, targetID)

	// Iterate over closest contacts and send FindData requests
	for _, contact := range closestContacts {
		messageBytes, err := json.Marshal(findDataMessage)
		if err != nil {
			return nil, fmt-Errorf("Failed to marshal FindDataMessage: %w", err)
		}

		// Using the SendAndWait function instead of direct send
		response := network.SendAndWait(contact.Address, RPC_FINDVAL, targetID[:], []bytes{})
		
		// Check if response is empty during the network communication failure
		if len(response) == 0 {
			fmt.print("No response or invalid response received from %s\n", contact.Address)
			continue // Skip to the next contact
		}

		// Parse the response to get the data
		var responseMessage FindDataResponseMessage
		err = json.Unmarshal(response, &responseMessage)
		if err != nil {
			fmt.Printf("Failed to unmarshal FindDataResponseMessage: %v\n", contact.Address, err)
			continue // Skip responses that cannot be parsed
		}

		// Check if the data was found
		if responseMessage.Found {
			return responseMessage.Data, nil // If the response contains data, return the found data
		} else {
			// If the target is a node, return the contact information
			fmt.Printf("Target %s is a node; returning contact information for node %s\n", targetID, contact.Address)
			return []byte(contact.Address), nil
		}
	}

	// If no data or node is found in any response
	return nil, fmt.Errorf("Data or node not found")
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) SendStoreMessage(meta *MessageMetadata, target_id []byte, value []byte) {
}
