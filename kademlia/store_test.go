package kademlia

import "testing"

func TestStore(t *testing.T) {
	test_store := NewStore()
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	var_1 := "1"
	success := test_store.Store(id, var_1)

	if !success {
		t.Error("Store is reporting that it alread is stored, despite being empty")
	}

	if !test_store.EntryExists(id) {
		t.Error("Value with given ID is not stored")
	}

	ret_1, success := test_store.GetEntry(id)
	if (ret_1 != var_1) || !success {
		t.Error("Store is not finding the value")
	}
}

func TestGetEntry(t *testing.T) {
	test_store := NewStore()
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("1111111100000000000000000000000000000000")
	val_1 := "val1"
	val_2 := "val2"

	_, success := test_store.GetEntry(id)
	if success {
		t.Error("GetEntry returns success when entry has not been added")
	}

	test_store.Store(id, val_1)
	ret_1, success := test_store.GetEntry(id)
	if !success {
		t.Error("GetEntry returns false after entry has been added")
	}
	if val_1 != ret_1 {
		t.Error("Values do not match")
	}

	_, success = test_store.GetEntry(id2)
	if success {
		t.Error("GetEntry returns success when entry has not been added")
	}

	test_store.Store(id2, val_2)
	ret_2, _ := test_store.GetEntry(id2)
	if (ret_2 == ret_1) || (ret_2 != val_2) {
		t.Error("Value mismatch")
	}
}

func TestEntryExists(t *testing.T) {
	test_store := NewStore()
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("1111111100000000000000000000000000000000")

	success := test_store.EntryExists(id)
	if success {
		t.Error("EntryExists returns success when entry has not been added")
	}

	test_store.Store(id, "")
	success = test_store.EntryExists(id)
	if !success {
		t.Error("EntryExists returns false after entry has been added")
	}

	success = test_store.EntryExists(id2)
	if success {
		t.Error("EntryExists returns success when entry has not been added")
	}
}

func TestNewEntry(t *testing.T) {
	key := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	val := "test"
	test_entry := NewStore().NewEntry(key, val)

	if test_entry.key != key {
		t.Error("Key mismatch")
	}

	if test_entry.value != val {
		t.Error("Val mismatch")
	}
}

func TestNewSto(t *testing.T) {
	test_store := NewStore()
	if len(test_store.entries) > 0 {
		t.Error("Store is not initiated with an empty list")
	}
}
