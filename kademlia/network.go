package kademlia

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Network struct {
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("CB")
}

func Listen(ip string, port int) {
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8008"
		}
		http.HandleFunc("/", HandleCallback)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()
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
