package main

import (
	"d7024e/kademlia"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	test_network := kademlia.NewNetwork("127.0.0.1", "9000", 2_000)

	bootstrap_id := "FFFFFFFF00000000000000000000000000000000"
	os.Setenv("PORT", "9001")
	os.Setenv("IS_BOOTSTRAP_NODE", "true")
	os.Setenv("BOOTSTRAP_PORT", "9001")
	os.Setenv("BOOTSTRAP_NODE_ID", bootstrap_id)

  fmt.Println("Starting bootstrap node...")

	go main()
  
  go test_network.Listen()
  time.Sleep(200 * time.Millisecond)
	test_network.JoinNetwork("127.0.0.1:" + os.Getenv("BOOTSTRAP_PORT"))
  fmt.Println("Bootstrap node started")
	resp := kademlia.Trim(test_network.SendPing(bootstrap_id))
	assert.Equal(t, "Ping response from "+bootstrap_id, resp)

	os.Setenv("IS_BOOTSTRAP_NODE", "false")

	const NR_NODES int = 20
	port := 9002
	var nodes [NR_NODES]*kademlia.Network

	for i := 0; i < NR_NODES; i++ {
		node := kademlia.NewNetwork("127.0.0.1", fmt.Sprintf("%d", port), 10_100 + i * 100)
		go node.Listen()
		// network.InitializeCLI()
		node.JoinNetwork("127.0.0.1:" + os.Getenv("BOOTSTRAP_PORT"))
		port++
		nodes[i] = node
		fmt.Printf("Node %d created\n", i+1)
	}

	// BEGIN test PING

	const MAX_TEST_TIME = 10
	ping_tests_done := false

	time.AfterFunc(MAX_TEST_TIME*time.Second, func() {
		if !ping_tests_done {
			fmt.Printf("Ping tests timed out\n")
			os.Exit(1)
		}
	})

	resp = kademlia.Trim(nodes[0].SendPing(bootstrap_id))
	assert.Equal(t, "Ping response from "+bootstrap_id, resp)

	nr_tests := 100

	for i := 0; i < nr_tests; i++ {
		n1 := rand.Intn(NR_NODES)
		n2 := rand.Intn(NR_NODES)
		for n2 == n1 {
			n2 = rand.Intn(NR_NODES)
		}
		fmt.Printf("Node %d pinging node %d\n", n1, n2)
		resp = kademlia.Trim(nodes[n1].SendPing(nodes[n2].GetID()))
		assert.Equal(t, "Ping response from "+nodes[n2].GetID(), resp)
	}

	ping_tests_done = true

	// END test PING

	fmt.Println("Done")
}
