package kademlia

import (
	"crypto/rand"
	"log"
)

// rnd is random component
// itr is iterative component WHICH WILL NEVER BE ZERO
// value is the two combined (most used)
type AuthUUID struct {
	value [2]byte
}

// Create a uuid instance from existing bytes
func NewAuthUUID(rnd byte, iter byte) AuthUUID {
	return AuthUUID{[2]byte{rnd, iter}}
}

// Generate a new auth uuid
// Uses a value that should be iterated every time port is opened
func GenerateAuthUUID(iter byte) AuthUUID {
	rnd := make([]byte, 1)
	_, err := rand.Read(rnd)
	if err != nil {
		log.Fatal(err)
	}
	return AuthUUID{[2]byte{rnd[0], iter}}
}
