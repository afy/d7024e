package kademlia

import (
  "syscall"
  "os"
  "fmt"
  // "io"
  // "bytes"
  "bufio"
  "strings"
)

func (network *Network) InitializeCLI() {
  pipePath := "/tmp/kademlia_pipe"
  respPath := "/tmp/kademlia_resp"
  syscall.Mkfifo(pipePath, 0666)
  syscall.Mkfifo(respPath, 0666)


  fpipe,_ := os.OpenFile(pipePath, syscall.O_RDWR, os.ModeNamedPipe)
  fresp,_ := os.OpenFile(respPath, syscall.O_WRONLY, os.ModeNamedPipe)
  defer fpipe.Close()
  defer fresp.Close()
  defer os.Remove(pipePath)
  defer os.Remove(respPath)

  reader := bufio.NewReader(fpipe)
  for {
    line, err := reader.ReadString('\n')
    line = strings.TrimSpace(line)
    if err != nil {
      continue 
    }
    cmd := strings.Split(line, " ")
    switch cmd[0] {
    case "put":
      // target := NewKademliaID(cmd[1])
      // network.SendStoreValueMessage(cmd[1], cmd[2])
      
    case "get":
      // network.SendFindValueMessage(cmd[1])
      
    case "exit":
      os.Exit(0)
    case "ping":
      fmt.Println("Pinging: " + cmd[1])
      target := NewKademliaID(cmd[1])
      closestContacts := network.routing_table.FindClosestContacts(target, 1)
      if len(closestContacts) == 0 {
        fmt.Println("No closest node found")
        fresp.WriteString("No closest node found\n")
        continue
      }
      closestNode := closestContacts[0]
      pingMessage := append([]byte{RPC_PING}, []byte(cmd[1])...) 
      resp := network.SendAndWait(closestNode.Address, RPC_PING, pingMessage, nil)
      fresp.WriteString(string(resp) + "\n")
      //network.SendPingMessage()
    case "print_id":
      fmt.Println(network.routing_table.me.ID)
      fresp.WriteString(network.routing_table.me.ID.String() + "\n")
    case "response:":
    default:
      fmt.Println("Invalid command: " + cmd[0])
    }
  } 
  fmt.Println("Exited CLI")
}
