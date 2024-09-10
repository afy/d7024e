package kademlia

import (
	"log"
	"strconv"
)

// Used to test an error and crash if it isnt nil.
// I know go doesn't like asserts so I made the func name VERY obvious.
// Sue me.
func AssertAndCrash(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Used to parse the port number from an ip address.
func ParsePortNumber(address string) int {
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
		return -1
	} else {
		n, err := strconv.Atoi(address[i+1 : sl])
		AssertAndCrash(err)
		return n
	}
}
