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

	if id == nil {
		t.Error("Expected a KademliaID instance, got nil")
		return
	}

	if *id != expected {
		t.Errorf("Expected KademliaID %x, got %x", expected, *id)
	}
}

// TestNewRandomKademliaID ensures it generates a valid random KademliaID.
func TestNewRandomKademliaID(t *testing.T) {
	id := NewRandomKademliaID()

	if id == nil {
		t.Error("Expected a KademliaID instance, got nil")
		return
	}

	if len(id) != IDLength {
		t.Errorf("Expected KademliaID length %d, got %d", IDLength, len(id))
	}
}

// TestCalcDistance confirms that the distance calculation via XOR works as expected.
func TestCalcDistance(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()

	distance := id1.CalcDistance(id2)

	// Calculate expected distance using XOR manually
	expectedDistance := KademliaID{}
	for i := 0; i < IDLength; i++ {
		expectedDistance[i] = id1[i] ^ id2[i]
	}

	if *distance != expectedDistance {
		t.Errorf("Expected distance %x, got %x", expectedDistance, *distance)
	}
}

// TestLess checks that the Less function works correctly between different KademliaIDs.
func TestLess(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()

	// Test that if id1 < id2, the result is true and vice versa
	if id1.Less(id2) == id2.Less(id1) {
		t.Errorf("Expected Less to return different results for %x and %x", *id1, *id2)
	}
}

// TestEquals checks that the Equals function works correctly between different KademliaIDs.
func TestEquals(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()
	id3 := *id1 // Create a copy of id1 for comparison

	if !id1.Equals(&id3) {
		t.Errorf("Expected %x to equal %x", *id1, id3)
	}

	if id1.Equals(id2) {
		t.Errorf("Expected %x not to equal %x", *id1, *id2)
	}
}
