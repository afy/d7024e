package main

import (
  "testing"
  "fmt"
  "os"
  "d7024e/kademlia"
  "github.com/stretchr/testify/assert"
  "math/rand"
  "time"
)

func TestMain(t *testing.T) {
  bootstrap_id := "FFFFFFFF00000000000000000000000000000000"
  os.Setenv("PORT", "9001")
  os.Setenv("IS_BOOTSTRAP_NODE", "true")
  os.Setenv("BOOTSTRAP_PORT", "9001")
  os.Setenv("BOOTSTRAP_NODE_ID", bootstrap_id)

  test_network := kademlia.NewNetwork("localhost", "9000")

  go main()

  test_network.JoinNetwork("localhost:" + os.Getenv("BOOTSTRAP_PORT")) 
  resp := kademlia.Trim(test_network.SendPingMessage(bootstrap_id))
  assert.Equal(t, "Ping response from " + bootstrap_id, resp)

  os.Setenv("IS_BOOTSTRAP_NODE", "false")

  const NR_NODES int = 10
  port := 9002
  var nodes [NR_NODES]*kademlia.Network

  for i := 0; i < NR_NODES; i++ {
    node := kademlia.NewNetwork("localhost", fmt.Sprintf("%d", port))
    go node.Listen()
    // network.InitializeCLI()
    node.JoinNetwork("localhost:" + os.Getenv("BOOTSTRAP_PORT"))
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

  resp = kademlia.Trim(nodes[0].SendPingMessage(bootstrap_id))
  assert.Equal(t, "Ping response from " + bootstrap_id, resp)

  nr_tests := 20

  for i := 0; i < nr_tests; i++ {
    n1 := rand.Intn(NR_NODES)
    n2 := rand.Intn(NR_NODES)
    for n2 == n1 {
      n2 = rand.Intn(NR_NODES)
    }
    fmt.Printf("Node %d pinging node %d\n", n1, n2)
    resp = kademlia.Trim(nodes[n1].SendPingMessage(nodes[n2].GetID()))
    assert.Equal(t, "Ping response from " + nodes[n2].GetID(), resp)
  }

  ping_tests_done = true

  // END test PING

  fmt.Println("Done")
}
