package kademlia

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
	network.SendNodeLookup(init_addr, network.routing_table.me.ID)
}

// SendPingMessage handles a PING request.
// If target is this node, send ping response to original requester.
// Otherwise, find the closest node and send a PING rpc to it.
func (network *Network) ManagePing(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	if target.Equals(network.routing_table.me.ID) {
		fmt.Printf("Responding to PING from %s\n", req_addr)
		network.SendResponse(aid, req_addr, RESP_PING_OK, nil)
		return
	}

	fmt.Printf("Target is not this node. Finding closest node to %s\n", target.String())
	closest_contacts := network.routing_table.FindClosestContacts(target, 1)

	if len(closest_contacts) == 0 {
		fmt.Printf("No closest node found")
		response := []byte(fmt.Sprintf("No closest node found"))
		network.SendResponse(aid, req_addr, RESP_PING_FAIL, response)
		return
	}

	closest := closest_contacts[0]
	network.routing_table.me.CalcDistance(target)
	closest.CalcDistance(target)
	if network.routing_table.me.Less(&closest) {
		fmt.Printf("No closer node found")
		response := []byte(fmt.Sprintf("No closer node found"))
		network.SendResponse(aid, req_addr, RESP_PING_FAIL, response)
		return
	}

	var response = make(byte_arr_list, 1)
	response[0] = []byte(target_node_id)
	resp := network.SendAndWait(closest.Address, RPC_PING, response)
	network.SendResponse(aid, req_addr, resp.Rpc, nil)
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
			network.SendResponse(aid, req_addr, RESP_STORE_EXISTS, nil)
			return
		}

		fmt.Printf("Adding entry to store: %s:%s, req from %s\n", value_id, value, req_addr)
		network.data_store.Store(target, value)
		network.SendResponse(aid, req_addr, RESP_STORE_OK, nil)
		return
	}

	var params = make(byte_arr_list, 2)
	params[0] = []byte(value_id)
	params[1] = []byte(value)
	response := network.SendAndWait(closest.Address, RPC_STORE, params)
	network.SendResponse(aid, req_addr, response.Rpc, nil)

}

// Same as findnode, but if the target is node, return a value instead.
func (network *Network) ManageFindData(aid *AuthID, req_addr string, value_id string) {
	target := NewKademliaID(value_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, ALPHA)
	if network.data_store.EntryExists(target) {
		fmt.Println("Value found")
		network.SendResponse(aid, req_addr, RESP_VALFOUND, []byte(network.data_store.GetEntry(target)))
		return
	}

	var contact_buffer bytes.Buffer
	encoder := gob.NewEncoder(&contact_buffer)
	err := encoder.Encode(closest_contacts)
	AssertAndCrash(err)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, RESP_CONTACTS, contact_buffer.Bytes())
}

// Get k closest nodes from k-buckets and return
func (network *Network) ManageFindContact(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, PARAM_K)
	var contact_buffer bytes.Buffer
	encoder := gob.NewEncoder(&contact_buffer)
	err := encoder.Encode(closest_contacts)
	AssertAndCrash(err)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, RESP_CONTACTS, contact_buffer.Bytes())
}

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

	switch resp.Rpc {
	case RESP_PING_OK:
		return fmt.Sprintf("Ping response from %s\n", target_node_id)
	case RESP_PING_FAIL:
		return fmt.Sprintf("Ping fail; %s\n", resp.Data[0])
	default:
		return fmt.Sprintf("ERR: %+v\n", resp)
	}
}

// Send a STORE RPC and return the status message string
func (network *Network) SendStore(value_key string, value []byte) string {
	target := NewKademliaID(value_key)
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

	switch resp.Rpc {
	case RESP_STORE_OK:
		return "Value has been stored in the network\n"
	case RESP_STORE_EXISTS:
		return "Value already exists\n"
	default:
		return fmt.Sprintf("ERR: %+v\n", resp)
	}
}

// Send a FINDVAL RPC and return the status message string
func (network *Network) SendFindValue(value_key string) string {
	target := NewKademliaID(value_key)
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

	switch resp.Rpc {
	case RESP_VALFOUND:
		return fmt.Sprintf("Value: %s\n", resp.Data[0])
	case RESP_CONTACTS:
		return fmt.Sprintf("%s\n", ParseContactList(resp.Data[0]))
	default:
		return fmt.Sprintf("ERR: %+v\n", resp)
	}
}

// Send a FIND_NODE rpc and return the status message strong
func (network *Network) SendFindContact(addr string, target_node_id *KademliaID) string {
	target := NewKademliaID(target_node_id.String())

	closest_contacts := network.routing_table.FindClosestContacts(target, 1)
	if len(closest_contacts) == 0 {
		fmt.Println("No closest node found")
		return "No closest node found\n"
	}

	closest_node := closest_contacts[0]
	var params = make(byte_arr_list, 1)
	params[0] = []byte(target.String())
	resp := network.SendAndWait(closest_node.Address, RPC_FINDVAL, params)

	switch resp.Rpc {
	case RESP_CONTACTS:
		return fmt.Sprintf("%s\n", ParseContactList(resp.Data[0]))
	default:
		return fmt.Sprintf("ERR: %+v\n", resp)
	}
}

// Used in join node,
func (network *Network) SendNodeLookup(init_addr string, target_node_id *KademliaID) {

}
