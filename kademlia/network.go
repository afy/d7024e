package kademlia

import (
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
	network.PerformNodeLookup(init_addr, network.routing_table.me.ID)
}

func (network *Network) PerformNodeLookup(init_addr string, target_node_id *KademliaID) {
	var params = make([][]byte, 1)
	params[0] = []byte(target_node_id.String())
	network.SendAndWait(init_addr, RPC_FINDCONTACT, params)
}

// SendPingMessage handles a PING request.
// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) ManagePingMessage(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	if target.Equals(network.routing_table.me.ID) {
		fmt.Printf("Responding to PING from %s\n", req_addr)
		var response = make([][]byte, 1)
		response[0] = []byte(fmt.Sprintf("Ping response from %s", target_node_id))
		network.SendResponse(aid, req_addr, response)
	} else {
		fmt.Printf("Target is not this node. Finding closest node to %s\n", target.String())
		closest_contacts := network.routing_table.FindClosestContacts(target, 1)

		if len(closest_contacts) == 0 {
			fmt.Printf("No closest node found")
			var response = make([][]byte, 1)
			response[0] = []byte(fmt.Sprintf("No closest node found"))
			network.SendResponse(aid, req_addr, response)
			return
		}

		closest_node := closest_contacts[0]
		var response = make([][]byte, 1)
		response[0] = []byte(target_node_id)
		resp := network.SendAndWait(closest_node.Address, RPC_PING, response)
		network.SendResponse(aid, req_addr, resp)
	}
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) ManageFindContactMessage(aid *AuthID, req_addr string, target_node_id string) {
	//target := NewKademliaID(target_node_id)
	//closest_contacts := network.routing_table.FindClosestContacts(target, ALPHA)
	//b := SerializeData(closest_contacts)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	var response = make([][]byte, 0)
	network.SendResponse(aid, req_addr, response)
}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) ManageFindDataMessage(aid *AuthID, req_addr string, value_id string) {
	target := GetValueID(value_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, ALPHA)

	for _, contact := range closest_contacts {
		fmt.Printf("Sending FindData request to: %s\n", contact.Address)
		params := make([][]byte, 1)
		params[0] = target[:]
		response := network.SendAndWait(contact.Address, RPC_FINDVAL, params)

		// Check if response is empty during the network communication failure
		if len(response) == 0 {
			fmt.Printf("No response or invalid response received from %s\n", contact.Address)
			continue // Skip to the next contact
		}

		// Check if the first byte of the response indicates data found
		if len(response) > 1 && response[0][0] == 0x01 {
			fmt.Printf("Data found for target %x at %s\n", value_id, contact.Address)
			network.SendResponse(aid, req_addr, response[1:])

		} else if len(response) > 1 && response[0][0] == 0x00 {
			fmt.Printf("Target %x is a node; continuing search.\n", value_id)
			continue
		}
	}

	fmt.Println(req_addr)
	network.SendResponse(aid, req_addr, nil)
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) ManageStoreMessage(aid *AuthID, req_addr string, value_id string, value string) {
	target := GetValueID(value_id)
	closest := network.routing_table.FindClosestContacts(target, 1)[0]
	network.routing_table.me.CalcDistance(target)
	closest.CalcDistance(target)

	if network.routing_table.me.Less(&closest) {
		if !network.data_store.EntryExists(target) {
			fmt.Printf("Adding entry to store: %s:%s, req from %s\n", value_id, value, req_addr)
			var response = make([][]byte, 1)
			response[0] = []byte(fmt.Sprintf("Stored value"))
			network.SendResponse(aid, req_addr, response)
			return
		}

		fmt.Printf("Entry already exists: %s:%s, req from %s\n", value_id, value, req_addr)
		var response = make([][]byte, 1)
		response[0] = []byte(fmt.Sprintf("Value is already stored in the network"))
		network.SendResponse(aid, req_addr, response)

	} else {
		var params = make([][]byte, 2)
		params[0] = []byte(value_id)
		params[1] = []byte(value)
		response := network.SendAndWait(closest.Address, RPC_PING, params)
		network.SendResponse(aid, req_addr, response)
	}
}

// PUT.
// Send a STORE RPC and return the status message string
func (network *Network) SendStoreValueMessage(value_id string, value []byte) string {
	target := GetValueID(value_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closestNode := closest_contacts[0]
	var params = make([][]byte, 1)
	params[0] = []byte(value_id)
	resp := network.SendAndWait(closestNode.Address, RPC_STORE, params)[0]
	return string(resp) + "\n"
}

// GET.
// Send a FINDVAL RPC and return the status message string
func (network *Network) SendFindValueMessage(value_id string) string {
	target := GetValueID(value_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closestNode := closest_contacts[0]
	var params = make([][]byte, 1)
	params[0] = []byte(value_id)
	resp := network.SendAndWait(closestNode.Address, RPC_FINDVAL, params)[0]
	return string(resp) + "\n"
}

// PING.
// Send a PING RPC to the network and return the status message string.
func (network *Network) SendPingMessage(target_node_id string) string {
	target := NewKademliaID(target_node_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closestNode := closest_contacts[0]
	var params = make([][]byte, 1)
	params[0] = []byte(target_node_id)
	resp := network.SendAndWait(closestNode.Address, RPC_PING, params)[0]
	return string(resp) + "\n"
}
