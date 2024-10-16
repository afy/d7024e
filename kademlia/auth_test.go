package kademlia

import (
  "testing"
  "fmt"
)

func TestNewAuthID(t *testing.T) {
	b1 := [20]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78,
	}
	a1 := NewAuthID(b1)
	if a1.value != b1 {
		t.Error("Input value is not set correctly")
	}
}

func TestGenerateRandomAuthID(t *testing.T) {
	a1 := GenerateRandomAuthID()
	a2 := GenerateRandomAuthID()
	if a1 == a2 {
		t.Error("Two random AuthIDs match; note that this has a *slim* chance of happening. Run test again to confirm")
	}
}

func TestEquals(t *testing.T) {
	b1 := [20]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78,
	}
	b2 := [20]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0xFF,
	}
	b3 := [20]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78,
	}

	a1 := AuthID{b1}
	a2 := AuthID{b2}
	a3 := AuthID{b3}

	if a1.Equals(a2) {
		t.Error("Invalid match still equals true")
	}
	if !a1.Equals(a3) {
		t.Error("Two equal values return false")
	}
	if !a3.Equals(a1) {
		t.Error("Two equal values return false")
	}
}

func TestString(t *testing.T) {
	b1 := [20]byte{
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
		0x12, 0x34, 0x56, 0x78,
	}
	valid := "1234567890abcdef1234567890abcdef12345678"
	a1 := AuthID{b1}
  fmt.Println(a1.String() + "\n" + valid)
	if a1.String() != valid {
		t.Error("String does not return correct result")
	}
}
