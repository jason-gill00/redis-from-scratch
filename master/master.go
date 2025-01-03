package master

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/jason-gill00/redis-from-scratch/client"
	"github.com/jason-gill00/redis-from-scratch/command"
	"github.com/jason-gill00/redis-from-scratch/persistence"
	"github.com/jason-gill00/redis-from-scratch/resp"
)

type Master struct {
	addr              string
	config            map[string]string
	clients           map[*client.Client]bool
	store             *persistence.Store
	msgChan           chan client.ClientMsg
	closeChan         chan net.Conn
	replicationConfig map[string]string
	replicas          []net.Conn
}

func NewMaster(replicationConfig map[string]string, store *persistence.Store, config map[string]string, port string) *Master {
	return &Master{
		replicationConfig: replicationConfig,
		addr:              fmt.Sprintf("0.0.0.0:%s", port),
		config:            config,
		store:             store,
		msgChan:           make(chan client.ClientMsg),
		closeChan:         make(chan net.Conn),
	}

}

func (m *Master) Start() {
	l, err := net.Listen("tcp", m.addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	// Seperate thread to read incomming messages from clients
	go m.clientHandler()

	// This is responsible for accepting new connections
	m.acceptLoop(l)
}

func (m *Master) clientHandler() {
	for {
		select {
		case clientMsg := <-m.msgChan:
			serializedCommandArrays, err := resp.RESPDeserializeCommand(string(clientMsg.Msg))
			if err != nil {
				slog.Error("Encountered error deserializing command", "err", err)
				continue
			}

			for _, serializedCommandArray := range serializedCommandArrays {
				response, err := command.CacheCommandHandler(serializedCommandArray, m.store, m.config, m.replicationConfig)
				if err != nil {
					slog.Error("Encountered error when handling command", "err", err)
					continue
				}
				m.write(response, clientMsg.Conn)

				// TODO: Find a better way to do this
				if strings.Contains(response, "FULLRESYNC") {
					emptyRDBhex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
					buf, _ := hex.DecodeString(emptyRDBhex)
					resp := resp.RESPSerializeFile(string(buf))
					m.write(resp, clientMsg.Conn)
				}

				// If the command is a PSYNC command and the server is a master, add the connection to the replicas list
				if serializedCommandArray[0] == "PSYNC" {
					m.replicas = append(m.replicas, clientMsg.Conn)
				}

				// If it is a write command, replicate to the slave
				if serializedCommandArray[0] == "SET" {
					for _, replConn := range m.replicas {
						fmt.Println("Replicating command to slavexxxx")
						m.write(string(clientMsg.Msg), replConn)
					}
				}
			}

		case closeConn := <-m.closeChan:
			slog.Info("Closing connection", "info", closeConn.RemoteAddr().String())
			closeConn.Close()
		}
	}
}

/*
* Responsible for accepting new connections and appending the connection to the clients map
 */
func (m *Master) acceptLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("error accepting connection", "err", err)
			continue
		}

		client := client.NewClient(conn, m.msgChan, m.closeChan)
		// m.clients[client] = true

		go client.ReadLoop()
	}
}

func (m *Master) write(response string, conn net.Conn) {
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error encountered when writing response: %s", err.Error())
	}
}
