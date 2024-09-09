package kademlia

import (
	"fmt"
	"log"
	"net/http"
	// "os"
)

type Network struct {
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("CB")
  fmt.Fprint(w, "Hello\n")
}

func Listen(ip string, port string) {
  fmt.Println("Starting server at port " + port) 
  http.HandleFunc("/", HandleCallback)
  err := http.ListenAndServe(ip+":"+port, nil)
  if err != nil {
    log.Fatal()
  }
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
