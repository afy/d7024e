package kademlia

// ===== REQ UDP PROTOCOL
// Byte 0:
//		0x0: PING
//		0x1: STORE
//		0x2: FINDNODE
//		0x3: FINDVAL
// Byte 1-2:
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
// Byte 2-END: Data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
)

// Send a request to the bootstrap node (init_addr) to join the network.
// This is done by requesting a self-lookup to the bootstrap node
func (network *Network) JoinNetwork(init_addr string) {
	fmt.Println("Self-lookup request sent")
	bootstrap_id := NewKademliaID(os.Getenv("BOOTSTRAP_NODE_ID"))
	network.routing_table.AddContact(NewContact(bootstrap_id, init_addr))

	// No response handling necessary, but we should still wait for it
	network.SendAndWait(init_addr, RPC_FINDCONTACT, []byte(os.Getenv("BOOTSTRAP_NODE_ID")), nil)
}

// SendPingMessage handles a PING request.
// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) ManagePingMessage(aid *AuthID, req_addr string, target_id string) {
	target := NewKademliaID(target_id)
	if target.Equals(network.routing_table.me.ID) {
		fmt.Printf("Responding to PING from %s\n", req_addr)
		response := []byte(fmt.Sprintf("Ping response from %s", target_id))
		network.SendResponse(aid, req_addr, response)
	} else {
		fmt.Printf("Target is not this node. Finding closest node to %s\n", target.String())
		closestContacts := network.routing_table.FindClosestContacts(target, 1)

		if len(closestContacts) == 0 {
			fmt.Printf("No closest node found")
			return
		}

		closestNode := closestContacts[0]
		resp := network.SendAndWait(closestNode.Address, RPC_PING, []byte(target_id), nil)
		network.SendResponse(aid, req_addr, resp)
	}
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) ManageFindContactMessage(aid *AuthID, req_addr string, target_id string) {
	targetID := NewKademliaID(target_id)
	closestContacts := network.routing_table.FindClosestContacts(targetID, ALPHA)
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(closestContacts)
	AssertAndCrash(err)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, b.Bytes())
}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) ManageFindDataMessage(aid *AuthID, req_addr string, target_id string) {
	targetID := NewKademliaID(target_id)
	closestContacts := network.routing_table.FindClosestContacts(targetID, ALPHA)

	for _, contact := range closestContacts {
		fmt.Printf("Sending FindData request to: %s\n", contact.Address)
		response := network.SendAndWait(contact.Address, RPC_FINDVAL, targetID[:], []byte{})

		// Check if response is empty during the network communication failure
		if len(response) == 0 {
			fmt.Printf("No response or invalid response received from %s\n", contact.Address)
			continue // Skip to the next contact
		}

		// Check if the first byte of the response indicates data found
		if len(response) > 1 && response[0] == 0x01 {
			fmt.Printf("Data found for target %x at %s\n", target_id, contact.Address)
			network.SendResponse(aid, req_addr, response[1:])

		} else if len(response) > 1 && response[0] == 0x00 {
			fmt.Printf("Target %x is a node; continuing search.\n", target_id)
			continue
		}
	}

	fmt.Println(req_addr)
	network.SendResponse(aid, req_addr, nil)
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) ManageStoreMessage(aid *AuthID, addr string, target_id []byte, value []byte) {
}

// PUT.
// Send a STORE RPC and return the status message string
func (network *Network) SendStoreValueMessage(target_id string, value []byte) string {
	return "Not implemented yet"
}

// GET.
// Send a FINDVAL RPC and return the status message string
func (network *Network) SendFindValueMessage(value_id string) string {
	return "Not implemented yet"
}

// PING.
// Send a PING RPC to the network and return the status message string.
func (network *Network) SendPingMessage(target_id string) string {
	target := NewKademliaID(string(target_id))
	closestContacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closestContacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closestNode := closestContacts[0]
	pingMessage := append([]byte{}, []byte(target_id)...)
	my_ip := network.routing_table.me.Address
	resp := network.SendAndWait(closestNode.Address, RPC_PING, pingMessage, []byte(my_ip))
	return string(resp) + "\n"
}
