package kademlia

import (
	"fmt"
	"log"
	"net/http"
	"net"
  "strings"
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
    log.Fatal(err)
  }
}

func ListenUDP(ip string, port string) {
  conn, err := net.ListenPacket("udp", ip+":"+port)
  
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()

  for {
    buf := make([]byte, 1024)
    n, addr, err := conn.ReadFrom(buf)
    if err != nil {
      log.Fatal(err)
      continue
    }
    
		data := strings.TrimSpace(string(buf[:n]))
		fmt.Printf("received: %s from %s\n", data, addr)
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
