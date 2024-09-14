package kademlia 

// Necessary imports
import (
	"errors"
	"time"
	"crypto/sha256"
	"fmt"
)

// TypeMessage for the Kademlia operation
type TypeMessage string

const (
	ERROR TypeMessage= "ERROR"
	STORE TypeMessage = "STORE"
	STORE_CONTACT TypeMessage = "STORE_CONTACT"
	FIND_CONTACT TypeMessage = "FIND_CONTACT"
	FIND_VALUE TypeMessage = "FIND_VALUE"
)

// Function to check if the TypeMessage is valid
func (typeMessage TypeMessage) IsValid() error {
	switch typeMessage {
	case ERROR, STORE, STORE_CONTACT, FIND_CONTACT, FIN_VALUE:
		return nil
	}
	return errors.New("Invalid type message")
}

// Store represents the key-value pair being stored
type Store struct {
	Key *Key
	Value string
}

// Kademlia key type
type Key struct {
	ID string
}

// Kademlia value-type with value and version
type Value struct {
	Value string
	Version int64
}

// DedupStore holds a value and all keys referencing it
type DedupStore struct {
	Value   string
	RefKeys []string
}

// Kademlia node with its datastore
type Node struct {
	dataStore map[string]Value  // Stores key-value with version
	dedupStore map[string]DedupStore  // Deduplication store
	valueHashes map[string]string  // Maps value hashes to keys for SavedStore 
}

// A new Kademlia node initialization 
func NewNode() *Node{
	return &Node{
		dataStore: make(map[string]Value),
		dedupStore: make(map[string]DedupStore),
		valueHashes: make(map[string]string),
	}
}

// Function to calculate hash value for SavedStore 
func hash(value string) string {
	h := sha256.New()
	h.write([]byte(value))
	return fmt.Sprintf("%x", h.sum(nil))
} 

// Function Store to handle different scenarios
func (n *Node) Store(storeRequest *Store) error {
	// Scenario 1: Handle same key but different value (versioning)
	if nodeEntry := n.dataStore[storeRequest.Key.ID];
	n.dataStore[storeRequest.Key.ID] = Value{
		Value: storeRequest.Value,
		Version: time.Now().Unix(),
	} 



	// Scenario 2: Handle different key but same value (deduplication)
}