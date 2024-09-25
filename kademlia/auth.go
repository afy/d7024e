package kademlia

import (
	"bytes"
	"crypto/rand"
	"log"
)

// A struct used to verify message integrity in RPC responses, since ports can be re-used.
type AuthID struct {
	value [2]byte
}

// Create a id instance from existing bytes.
func NewAuthID(rnd byte, iter byte) AuthID {
	return AuthID{[2]byte{rnd, iter}}
}

// Generate a new auth id.
// Uses a value that should be iterated every time port is opened.
func GenerateAuthID(iter byte) AuthID {
	rnd := make([]byte, 1)
	_, err := rand.Read(rnd)
	if err != nil {
		log.Fatal(err)
	}
	return AuthID{[2]byte{rnd[0], iter}}
}

// Compare two ids for equality.
func (auth_id *AuthID) Equals(a AuthID) bool {
	return bytes.Equal(auth_id.value[:], a.value[:])
}
