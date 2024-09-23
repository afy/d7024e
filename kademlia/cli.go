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
  syscall.Mkfifo(pipePath, 0666)


  fpipe,_ := os.OpenFile(pipePath, syscall.O_RDWR, os.ModeNamedPipe)
  defer fpipe.Close()
  defer os.Remove(pipePath)

  reader := bufio.NewReader(fpipe)
  for {
    line, err := reader.ReadString('\n')
    line = strings.TrimSpace(line)
    if err != nil {
      break
    }
    fmt.Printf("Received: %s\n", line)
  } 
}
