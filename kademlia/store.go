package kademlia

// Necessary imports
import (
	"log"
)

type Entry struct {
	key   *KademliaID
	value string
}

type Store struct {
	entries []*Entry
}

func NewStore() *Store {
	var s []*Entry
	return &Store{s}
}

func (store *Store) NewEntry(hash *KademliaID, value string) *Entry {
	return &Entry{hash, value}
}

func (store *Store) Store(hash *KademliaID, value string) {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			log.Println("Value is already stored")
			return
		}
	}
	ne := store.NewEntry(hash, value)
	store.entries = append(store.entries, ne)
}

func (store *Store) GetEntry(hash *KademliaID) string {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			return e.value
		}
	}
	log.Println("Value is not stored")
	return "NIL"
}

func (store *Store) EntryExists(hash *KademliaID) bool {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			return true
		}
	}
	return false
}
