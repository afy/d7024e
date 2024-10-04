package main

import (
	"d7024e/kademlia"
	"fmt"
	"os"
	"strconv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8008"
	}
	net := kademlia.NewNetwork("0.0.0.0", port)
	go net.Listen()
	go net.InitializeCLI()

	is_bootstrap, err := strconv.ParseBool(os.Getenv("IS_BOOTSTRAP_NODE"))
	kademlia.AssertAndCrash(err)

	if !is_bootstrap {
		fmt.Println("Attempting to join network...")
		net.JoinNetwork("bootstrap-node:" + os.Getenv("BOOTSTRAP_PORT"))
	} else {
    fmt.Printf("Running bootstrap node on port %s\n", port)
  }

	for {
		kademlia.UpdateTimers()
	}
}
