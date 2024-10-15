package kademlia

import (
	"fmt"
	"os"
	"sort"
)

// Send a request to the bootstrap node (init_addr) to join the network.
// This is done by requesting a self-lookup to the bootstrap node
func (network *Network) JoinNetwork(init_addr string) {
	fmt.Println("Self-lookup request sent")
	bootstrap_id := NewKademliaID(os.Getenv("BOOTSTRAP_NODE_ID"))
	network.routing_table.AddContact(NewContact(bootstrap_id, init_addr))

	var params = make(byte_arr_list, 1)
	target_node_id := network.routing_table.me.ID.String()
	params[0] = []byte(target_node_id)
	resp := network.SendAndWait(init_addr, RPC_NODELOOKUP, params)

	// Send ping to nodes
	fmt.Println(ParseContactList(resp.Data[0]))
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
	closest_contacts := network.routing_table.FindClosestContacts(target, PARAM_K)
	fmt.Printf("%+v\n", closest_contacts)

	if network.data_store.EntryExists(target) {
		fmt.Println("Value found")
		val, _ := network.data_store.GetEntry(target)
		network.SendResponse(aid, req_addr, RESP_VALFOUND, []byte(val))
		return
	}

	contact_bytes := NetSerialize[[]Contact](closest_contacts)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, RESP_CONTACTS, contact_bytes)
}

// Get k closest nodes from k-buckets and return
func (network *Network) ManageFindContact(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	closest_contacts := network.routing_table.FindClosestContacts(target, PARAM_K)
	contact_bytes := NetSerialize[[]Contact](closest_contacts)
	fmt.Printf("Main listener: Sent response to: %s\n", req_addr)
	network.SendResponse(aid, req_addr, RESP_CONTACTS, contact_bytes)
}

func (network *Network) ManageNodeLookup(aid *AuthID, req_addr string, target_node_id string) {
	target := NewKademliaID(target_node_id)
	closest := network.routing_table.FindClosestContacts(target, ALPHA)
	var params = make(byte_arr_list, 1)
	params[0] = []byte(target_node_id)

	first_pass_ch := make(chan []Contact, ALPHA)
	recursion_result := make(chan []Contact, ALPHA)
	var shortlist []Contact

	// First-pass: Send alpha requests and initiate second (recursive) step when they return
	for _, c := range closest {
		go func() {
			resp := network.SendAndWait(c.Address, RPC_FINDCONTACT, params)
			first_pass_ch <- NetDeserialize[[]Contact](resp.Data[0])
		}()
	}

	// recursive case
	for _, _ = range closest {
		x := <-first_pass_ch
		go func(unqueried []Contact) {
			var ret []Contact
			for len(unqueried) > 0 {
				// prevent loop by sending FC to self
				if !(unqueried[0].ID.Equals(network.routing_table.me.ID)) {
					resp := network.SendAndWait(unqueried[0].Address, RPC_FINDCONTACT, params)
					ret = append(ret, NetDeserialize[[]Contact](resp.Data[0])...)
				}
				unqueried = unqueried[1:]
			}
			fmt.Printf("a- %+v\n", shortlist)
			fmt.Printf("b- %+v\n", unqueried)
			recursion_result <- ret
		}(x)
	}

	// Wait for recursion steps to finish
	for _, _ = range closest {
		res := <-recursion_result
		shortlist = append(shortlist, res...)
		fmt.Printf("%+v\n", shortlist)
		sort.Slice(shortlist, func(i, j int) bool {
			return shortlist[i].Less(&shortlist[j])
		})

		if len(shortlist) > 20 {
			shortlist = shortlist[:20]
		}
	}

	contact_bytes := NetSerialize[[]Contact](shortlist)
	network.SendResponse(aid, req_addr, RESP_CONTACTS, contact_bytes)
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
		return "Value has been stored in the network\n"
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
		val, _ := network.data_store.GetEntry(target)
		return val + "\n"
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
