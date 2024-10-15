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

func (store *Store) Store(hash *KademliaID, value string) bool {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			log.Println("Value is already stored")
			return false
		}
	}
	ne := store.NewEntry(hash, value)
	store.entries = append(store.entries, ne)
	return true
}

func (store *Store) GetEntry(hash *KademliaID) (string, bool) {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			return e.value, true
		}
	}
	log.Println("Value is not stored")
	return "", false
}

func (store *Store) EntryExists(hash *KademliaID) bool {
	for _, e := range store.entries {
		if e.key.Equals(hash) {
			return true
		}
	}
	return false
}
