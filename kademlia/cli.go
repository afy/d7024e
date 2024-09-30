package kademlia

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func (network *Network) InitializeCLI() {
	pipe_path := "/tmp/kademlia_pipe"
	resp_path := "/tmp/kademlia_resp"
	syscall.Mkfifo(pipe_path, 0666)
	syscall.Mkfifo(resp_path, 0666)

	fpipe, _ := os.OpenFile(pipe_path, syscall.O_RDWR, os.ModeNamedPipe)
	fresp, _ := os.OpenFile(resp_path, syscall.O_WRONLY, os.ModeNamedPipe)
	defer fpipe.Close()
	defer fresp.Close()
	defer os.Remove(pipe_path)
	defer os.Remove(resp_path)

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
			status := network.SendStore(cmd[1], []byte(cmd[2]))
			fresp.WriteString(status)

		case "get":
			status := network.SendFindValue(cmd[1])
			fresp.WriteString(status)

		case "exit":
			fresp.WriteString("Exiting...\n")
			os.Exit(0)

		case "ping":
			status := network.SendPing(cmd[1])
			fresp.WriteString(status)

		case "print_id":
			fmt.Println(network.routing_table.me.ID)
			fresp.WriteString(network.routing_table.me.ID.String() + "\n")

		case "response:":
		default:
			fmt.Println("Invalid command: " + cmd[0])
			fresp.WriteString("Invalid command\n")
		}
	}
}
