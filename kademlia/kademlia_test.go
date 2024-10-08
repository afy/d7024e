package kademlia

import (
	"testing"
)

// TestNewKademliaID verifies that a valid string input produces a correct KademliaID.
func TestNewKademliaID(t *testing.T) {
	input := "1234567890abcdef1234567890abcdef12345678"
	expected := KademliaID{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78,
	}

	id := NewKademliaID(input)
	t.Logf("Testing NewKademliaID with input: %s, got: %x\n", input, *id)

	if id == nil {
		t.Error("Expected a KademliaID instance, got nil")
		return
	}

	if *id != expected {
		t.Errorf("Expected KademliaID %x, got %x", expected, *id)
	}
}

// TestKademliaIDLess checks that the Less function works correctly between different KademliaIDs.
func TestKademliaIDLess(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()

	t.Logf("Testing Less function: id1: %x, id2: %x\n", *id1, *id2)

	// Test that if id1 < id2, the result is true and vice versa
	if id1.Less(id2) == id2.Less(id1) {
		t.Errorf("Expected Less to return different results for %x and %x", *id1, *id2)
	}
}

// TestKademliaIDComparison verifies that KademliaID comparisons work as expected.
func TestKademliaIDComparison(t *testing.T) {
	// Using valid hex strings to create predictable KademliaIDs
	id1 := NewKademliaID("0000000000000000000000000000000000000000") // Lowest value
	id2 := NewKademliaID("0000000000000000000000000000000000000001") // Just higher than id1
	id3 := NewKademliaID("0000000000000000000000000000000000000000") // Same as id1

	t.Logf("Comparing KademliaIDs: id1: %x, id2: %x, id3: %x\n", *id1, *id2, *id3)

	// Test Equals
	if !id1.Equals(id3) {
		t.Error("Expected id1 to equal id3")
	}

	if id1.Equals(id2) {
		t.Error("Expected id1 to not equal id2")
	}

	// Test Less
	if !id1.Less(id2) {
		t.Error("Expected id1 to be less than id2")
	}

	if id2.Less(id1) {
		t.Error("Expected id2 to not be less than id1")
	}
}

// TestKademliaIDEquals verifies that equality is correctly determined between different KademliaIDs.
func TestKademliaIDEquals(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()
	id3 := *id1 // Create a copy of id1 for comparison

	t.Logf("Testing equality: id1: %x, id2: %x, id3: %x\n", *id1, *id2, id3)

	if !id1.Equals(&id3) {
		t.Errorf("Expected %x to equal %x", *id1, id3)
	}

	if id1.Equals(id2) {
		t.Errorf("Expected %x not to equal %x", *id1, *id2)
	}
}
