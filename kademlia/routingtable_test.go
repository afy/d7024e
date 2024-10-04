package kademlia

import (
	"fmt"
	"testing"
  "github.com/stretchr/testify/assert"
)

// FIXME: This test doesn't actually test anything. There is only one assertion
// that is included as an example.

func TestRoutingTable(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002"))

	contacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	for i := range contacts {
		fmt.Println(contacts[i].String())
	}

	// TODO: This is just an example. Make more meaningful assertions.
  assert.Equal(t, 6, len(contacts), fmt.Sprintf("Expected 6 contacts but instead got %d", len(contacts)))
}

// TestFindClosestContacts ensures that it retrieves the expected number of closest contacts.
func TestFindClosestContacts(t *testing.T) {
	// Initialize with a local contact
	localContact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(localContact)

	contact1 := NewContact(NewKademliaID("1234567890abcdef1234567890abcdef1234"), "192.168.1.1:8000")
	contact2 := NewContact(NewKademliaID("abcdef1234567890abcdef12345678901234"), "192.168.1.2:8001")
	contact3 := NewContact(NewKademliaID("11223344556677889900aabbccddeeff0011"), "192.168.1.3:8002")

	rt.AddContact(contact1)
	rt.AddContact(contact2)
	rt.AddContact(contact3)

	targetID := NewKademliaID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	closestContacts := rt.FindClosestContacts(targetID, 2)

	// Check if the number of closest contacts is correct
	if len(closestContacts) != 2 {
		t.Errorf("Expected 2 closest contacts, but got: %d", len(closestContacts))
	}

	// Optionally check if the closest contacts are as expected
	// You can implement specific checks based on your distance calculations.
}

// TestGetBucketIndex verifies that the correct bucket index is returned based on the KademliaID.
func TestGetBucketIndex(t *testing.T) {
	// Initialize with a local contact
	localContact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(localContact)

	// This should return the bucket index based on the distance of the KademliaID
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	bucketIndex := rt.getBucketIndex(id)

	expectedIndex := 0 // Adjust based on your bucket indexing logic
	if bucketIndex != expectedIndex {
		t.Errorf("Expected bucket index %d, got %d", expectedIndex, bucketIndex)
	}
}
