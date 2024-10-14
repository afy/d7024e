package kademlia

// This file contains helper functions/constants that dont directly
// depend on any other structs.

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Used to test an error and crash if it isnt nil.
// I know go doesn't like asserts so I made the func name VERY obvious.
// Sue me.
func AssertAndCrash(err error) {
	if err != nil {
		panic(err)
	}
}

// Used to parse the port number from an ip address.
func ParsePortNumber(address string) (string, int) {
	var i int
	var c rune
	for i, c = range address {
		if string(c) == ":" {
			break
		}
	}

	sl := len(address)
	if i == sl {
		log.Fatal("Cannot parse port from address: " + address)
		return "", -1
	} else {
		n, err := strconv.Atoi(address[i+1 : sl])
		AssertAndCrash(err)
		return address[0:i], n
	}
}

// Translate RPC code (byte) to a string for printing
func GetRPCName(code byte) string {
	switch code {
	case RPC_PING:
		return "PING"
	case RPC_STORE:
		return "STORE"
	case RPC_FINDVAL:
		return "FINDVAL"
	case RPC_FINDCONTACT:
		return "FINDCONTACT"
	case RPC_NODELOOKUP:
		return "NODELOOKUP"
	default:
		return "[ERR]"
	}
}

// data string value -> kademlia id
func GetValueID(val string) *KademliaID {
	sha := sha1.New()
	sha.Write([]byte(val))
	sum := sha.Sum(nil)
	s := fmt.Sprintf("%x", sum)
	return NewKademliaID(s)
}

// Format contact list to printable string
func ParseContactList(raw []byte) string {
	buff := bytes.NewBuffer(raw)
	data := NetDeserialize[[]Contact](buff.Bytes())
	ret := ""
	for _, e := range data {
		line := fmt.Sprintf("<%s, %s>", e.Address, e.ID.String())
		ret = ret + line
	}
	return ret
}

func Trim(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\x00", "")
}

func NetSerialize[T any](data any) []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	AssertAndCrash(err)
	return buff.Bytes()
}

func NetDeserialize[T any](data []byte) T {
	byte_buffer := bytes.NewBuffer(data)
	var ret T
	decoder := gob.NewDecoder(byte_buffer)
	err := decoder.Decode(&ret)
	AssertAndCrash(err)
	return ret
}
