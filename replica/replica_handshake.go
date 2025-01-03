package replica

import (
	"fmt"
	"net"
	"strings"

	"github.com/jason-gill00/redis-from-scratch/client"
)

func sendCommand(conn net.Conn, command string, expectedResponse string) error {
	_, err := conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("error sending command: %s", err)
	}
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("error reading response: %s", err)
	}
	if !strings.Contains(string(response[:n]), expectedResponse) {
		return fmt.Errorf("unexpected response from master: %s", string(response[:n]))
	}
	return nil
}

func InitiateHandshake(masterAddr, port string, msgChan chan client.ClientMsg, closeChan chan net.Conn) (net.Conn, error) {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to master: %s", err)
	}
	// defer conn.Close()
	pingCommand := "*1\r\n$4\r\nPING\r\n"
	if err := sendCommand(conn, pingCommand, "PONG"); err != nil {
		return nil, fmt.Errorf("PING failed: %s", err)
	}
	replConfPort := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%s\r\n", port)
	if err := sendCommand(conn, replConfPort, "OK"); err != nil {
		return nil, fmt.Errorf("REPLCONF listening-port failed: %s", err)
	}
	replConfCapa := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if err := sendCommand(conn, replConfCapa, "OK"); err != nil {
		return nil, fmt.Errorf("REPLCONF capa failed: %s", err)
	}

	client := client.NewClient(conn, msgChan, closeChan)
	go client.ReadLoop()

	psync := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	_, err = conn.Write([]byte(psync))
	if err != nil {
		return nil, fmt.Errorf("error sending command: %s", err)
	}
	// if err := sendCommand(conn, psync, "FULLRESYNC"); err != nil {
	// 	return nil, fmt.Errorf("PSYNC failed: %s", err)
	// }

	return conn, nil
}
