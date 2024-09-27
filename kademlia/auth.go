package kademlia

import (
	"bytes"
	"crypto/rand"
	"log"
)

// A struct used to verify message integrity in RPC responses, since ports can be re-used.
type AuthID struct {
	value [20]byte
}

// Create a id instance from existing bytes.
func NewAuthID(d [20]byte) *AuthID {
	return &AuthID{d}
}

// Generate a new auth id.
func GenerateAuthID() *AuthID {
	rnd := make([]byte, 20)
	_, err := rand.Read(rnd)
	if err != nil {
		log.Fatal(err)
	}
	var d [20]byte
	copy(rnd[:], d[:20])
	return &AuthID{d}
}

// Compare two ids for equality.
func (auth_id *AuthID) Equals(a AuthID) bool {
	return bytes.Equal(auth_id.value[:], a.value[:])
}

func (auth_id *AuthID) String() string {
	return string(auth_id.value[:])
}
