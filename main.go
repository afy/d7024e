package main

import (
  "os"
  "d7024e/kademlia"
)

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8008"
  }
  kademlia.Listen("", port)
}
