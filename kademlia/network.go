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
	network.SendFindContact(init_addr, network.routing_table.me.ID)
}

// SendPingMessage handles a PING request.
// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) ManagePing(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	if target.Equals(network.routing_table.me.ID) {
		fmt.Printf("Responding to PING from %s\n", req_addr)
		response := []byte(fmt.Sprintf("Ping response from %s", target_node_id))
		network.SendResponse(aid, req_addr, response)
	} else {
		fmt.Printf("Target is not this node. Finding closest node to %s\n", target.String())
		closest_contacts := network.routing_table.FindClosestContacts(target, 1)

		if len(closest_contacts) == 0 {
			fmt.Printf("No closest node found")
			response := []byte(fmt.Sprintf("No closest node found"))
			network.SendResponse(aid, req_addr, response)
			return
		}

		closest_node := closest_contacts[0]
		var response = make(byte_arr_list, 1)
		response[0] = []byte(target_node_id)
		resp := network.SendAndWait(closest_node.Address, RPC_PING, response)
		network.SendResponse(aid, req_addr, resp)
	}
}

// Same as PING but send additional metadata that gets stored. Send an OK to original client.
func (network *Network) ManageStore(aid *AuthID, req_addr string, value_id string, value string) {
	target := NewKademliaID(value_id)
	closest := network.routing_table.FindClosestContacts(target, 1)[0]

	network.routing_table.me.CalcDistance(target)
	closest.CalcDistance(target)
	if network.routing_table.me.Less(&closest) {
		if network.data_store.EntryExists(target) {
			fmt.Printf("Entry already exists: %s:%s, req from %s\n", value_id, value, req_addr)
			response := []byte(fmt.Sprintf("Value is already stored in the network"))
			network.SendResponse(aid, req_addr, response)
		} else {
			fmt.Printf("Adding entry to store: %s:%s, req from %s\n", value_id, value, req_addr)
			network.data_store.Store(target, value)
			response := []byte(fmt.Sprintf("Stored value %s : %s at node %s", value_id, value, network.routing_table.me.ID.String()))
			network.SendResponse(aid, req_addr, response)
		}

	} else {
		var params = make(byte_arr_list, 2)
		params[0] = []byte(value_id)
		params[1] = []byte(value)
		response := network.SendAndWait(closest.Address, RPC_STORE, params)
		network.SendResponse(aid, req_addr, response)
	}
}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) ManageFindData(aid *AuthID, req_addr string, value_id string) {
	target := NewKademliaID(value_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, ALPHA)

	if network.data_store.EntryExists(target) {
		fmt.Println("Value found")
		network.SendResponse(aid, req_addr, []byte(network.data_store.GetEntry(target)))
	}

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
		if len(response) > 1 && response[0] == 0x01 {
			fmt.Printf("Data found for target %x at %s\n", value_id, contact.Address)
			network.SendResponse(aid, req_addr, response[1:])

		} else if len(response) > 1 && response[0] == 0x00 {
			fmt.Printf("Target %x is a node; continuing search.\n", value_id)
			continue
		}
	}

	fmt.Println(req_addr)
	network.SendResponse(aid, req_addr, nil)
}

// Get "alpha" closest nodes from k-buckets and send simultaneous reqs.
// collect a list of the k-closest nodes and send back to client.
func (network *Network) ManageFindContact(aid *AuthID, req_addr string, target_node_id string) {
	//target := NewKademliaID(target_node_id)
	//closest_contacts := network.routing_table.FindClosestContacts(target, ALPHA)
	//b := SerializeData(closest_contacts)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, nil)
}

// PING.
// Send a PING RPC to the network and return the status message string.
func (network *Network) SendPing(target_node_id string) string {
	target := NewKademliaID(target_node_id)
	if network.routing_table.me.ID.Equals(target) {
		fmt.Println("Ping is this node")
		return fmt.Sprintf("Ping response from %s\n", target_node_id)
	}

	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closest_node := closest_contacts[0]
	var params = make(byte_arr_list, 1)
	params[0] = []byte(target_node_id)
	resp := network.SendAndWait(closest_node.Address, RPC_PING, params)
	return string(resp) + "\n"
}

// PUT.
// Send a STORE RPC and return the status message string
func (network *Network) SendStore(value_key string, value []byte) string {
	target := GetValueID(value_key)
	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}

	closest_node := closest_contacts[0]
	network.routing_table.me.CalcDistance(target)
	closest_node.CalcDistance(target)
	if network.routing_table.me.Less(&closest_node) {
		fmt.Printf("Adding entry to store: %s:%s, at self\n", target.String(), value)
		network.data_store.Store(target, string(value))
		return "Stored value at self\n"
	}

	var params = make(byte_arr_list, 2)
	params[0] = []byte(target.String())
	params[1] = []byte(value)
	resp := network.SendAndWait(closest_node.Address, RPC_STORE, params)
	return string(resp) + "\n"
}

// GET.
// Send a FINDVAL RPC and return the status message string
func (network *Network) SendFindValue(value_key string) string {
	target := GetValueID(value_key)
	if network.data_store.EntryExists(target) {
		fmt.Println("Value found")
		return network.data_store.GetEntry(target) + "\n"
	}

	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}
	closest_node := closest_contacts[0]
	var params = make(byte_arr_list, 1)
	params[0] = []byte(target.String())
	resp := network.SendAndWait(closest_node.Address, RPC_FINDVAL, params)
	return string(resp) + "\n"
}

// Used in join node, performs a FIND_NODE rpc
func (network *Network) SendFindContact(init_addr string, target_node_id *KademliaID) {
	var params = make(byte_arr_list, 1)
	params[0] = []byte(target_node_id.String())
	network.SendAndWait(init_addr, RPC_FINDCONTACT, params)
	return
}
