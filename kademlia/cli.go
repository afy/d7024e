package kademlia

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func (network *Network) InitializeCLI() {
	pipePath := "/tmp/kademlia_pipe"
	respPath := "/tmp/kademlia_resp"
	syscall.Mkfifo(pipePath, 0666)
	syscall.Mkfifo(respPath, 0666)

	fpipe, _ := os.OpenFile(pipePath, syscall.O_RDWR, os.ModeNamedPipe)
	fresp, _ := os.OpenFile(respPath, syscall.O_WRONLY, os.ModeNamedPipe)
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
			status := network.SendStoreValueMessage(cmd[1], []byte(cmd[2]))
			fresp.WriteString(status)

		case "get":
			status := network.SendFindValueMessage(cmd[1])
			fresp.WriteString(status)

		case "exit":
			fresp.WriteString("Exiting...\n")
			os.Exit(0)

		case "ping":
			status := network.SendPingMessage(cmd[1])
			fresp.WriteString(status)

		case "print_id":
			fmt.Println(network.routing_table.me.ID)
			fresp.WriteString(network.routing_table.me.ID.String() + "\n")

		case "response:":
		default:
			fmt.Println("Invalid command: " + cmd[0])
		}
	}
}
