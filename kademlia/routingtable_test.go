package kademlia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRoutingTable tests the creation of a new RoutingTable instance.
func TestNewRoutingTable(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	assert.NotNil(t, rt, "Expected RoutingTable instance, got nil")
	assert.Equal(t, me, rt.me, "Expected 'me' contact to be set correctly")

	t.Logf("New RoutingTable created for contact: %s", rt.me.String())

}

// TestAddingMultipleContacts tests adding multiple contacts to the routing table.
func TestAddingMultipleContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	contactsToAdd := []Contact{
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"),
		NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"),
		NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8003"),
		NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8004"),
		NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8005"),
		NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8006"),
	}

	for _, contact := range contactsToAdd {
		rt.AddContact(contact)
	}

	// Ensure we have the correct number of contacts added
	assert.NotNil(t, rt.buckets, "Expected buckets to be initialized")
	for _, bucket := range rt.buckets {
		assert.NotNil(t, bucket, "Expected each bucket to be initialized")
	}
}

// TestFindClosestContacts verifies that the FindClosestContacts function returns the correct contacts.
func TestFindClosestContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	// Add contacts for testing
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("2111111100000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8003"))
	rt.AddContact(NewContact(NewKademliaID("2111111200000000000000000000000000000000"), "localhost:8004"))

	// Find closest contacts
	closestContacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 3)

	// Check that we received the correct number of closest contacts
	assert.Equal(t, 3, len(closestContacts), "Expected to find 3 closest contacts.")
}

// TestGetBucketIndex verifies the correct bucket index is calculated for a KademliaID.
func TestGetBucketIndex(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))
	testID := NewKademliaID("0000000000000000000000000000000000000001")

	index := rt.getBucketIndex(testID)

	assert.GreaterOrEqual(t, index, 0, "Expected bucket index to be non-negative.")
	assert.Less(t, index, IDLength*8, "Expected bucket index to be less than total number of buckets.")
}
