package main

import (
	"d7024e/kademlia"
	"log"
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

	is_bootstrap, err := strconv.ParseBool(os.Getenv("IS_BOOTSTRAP_NODE"))
	if err != nil {
		log.Fatal(err)
	}

	if !is_bootstrap {
		net.JoinNetwork("bootstrap-node" + os.Getenv("BOOTSTRAP_PORT"))
	}

	kademlia.UpdateTimers()
}
