package replica

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/jason-gill00/redis-from-scratch/client"
	com "github.com/jason-gill00/redis-from-scratch/command"
	"github.com/jason-gill00/redis-from-scratch/persistence"
	"github.com/jason-gill00/redis-from-scratch/resp"
)

type command struct {
	command    []string
	rawCommand []byte
	conn       net.Conn
}

type Replica struct {
	commandBuffer     []command
	msgChan           chan client.ClientMsg
	closeChan         chan net.Conn
	addr              string
	config            map[string]string
	replicationConfig map[string]string
	store             *persistence.Store
	masterAddr        string
	offset            int
	port              string
}

func NewReplica(replicationConfig map[string]string, store *persistence.Store, config map[string]string, port string) *Replica {
	masterAddr := strings.ReplaceAll(replicationConfig["replicaof"], " ", ":")
	return &Replica{
		replicationConfig: replicationConfig,
		addr:              fmt.Sprintf("0.0.0.0:%s", port),
		masterAddr:        masterAddr,
		config:            config,
		store:             store,
		msgChan:           make(chan client.ClientMsg),
		closeChan:         make(chan net.Conn),
	}

}

// Process commands from the command buffer
func (r *Replica) readLoop() {
	for {
		if len(r.commandBuffer) == 0 {
			// sleep
			continue
		}
		// time.Sleep(1 * time.Second)
		processedCommand := r.commandBuffer[0]
		r.commandBuffer = r.commandBuffer[1:]

		// For now if it is a full resync command, we will ignore it
		if strings.Contains(processedCommand.command[0], "FULLRESYNC") {
			fmt.Println("Ignoring full resync command")
			continue
		}

		// If the command contains the rdb file, we will ignore it
		if strings.Contains(processedCommand.command[0], "REDIS0011") {
			fmt.Println("Ignoring rdb file")
			continue
		}

		response, err := com.CacheCommandHandler(processedCommand.command, r.store, r.config, r.replicationConfig)
		if err != nil {
			slog.Error("Encountered error when handling command", "err", err)
			continue
		}

		// Only write the response back to master if the command is REPLCONF
		if processedCommand.command[0] == "PING" {
			continue
		}
		fmt.Println("Writing response back to master")
		r.write(response, processedCommand.conn)

		rawCommand := resp.RESPSerializeRESPArray(processedCommand.command)

		// Update the offset. The offset is how many bytes we have read from the master
		r.offset += len(rawCommand)
		r.replicationConfig["slave_repl_offset"] = fmt.Sprintf("%d", r.offset)

	}
}

func (r *Replica) Start() {
	l, err := net.Listen("tcp", r.addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	err = r.preformMasterHandshake()
	if err != nil {
		slog.Error("Error preforming master handshake", "err", err)
	}

	go r.readLoop()

	// Seperate thread to read incomming messages from clients
	go r.clientHandler()

	// This is responsible for accepting new connections
	r.acceptLoop(l)
}

func (r *Replica) preformMasterHandshake() error {
	_, err := InitiateHandshake(r.masterAddr, r.port, r.msgChan, r.closeChan)
	if err != nil {
		return err
	}
	return nil
}

func (r *Replica) clientHandler() {
	for {
		select {
		case clientMsg := <-r.msgChan:
			slog.Info("Received msg from chan", "info", string(clientMsg.Msg))
			serializedCommandArrays, err := resp.RESPDeserializeCommand(string(clientMsg.Msg))
			if err != nil {
				slog.Error("Encountered error deserializing command", "err", err)
				continue
			}

			for _, serializedCommandArray := range serializedCommandArrays {
				fmt.Println("Appending command: ", serializedCommandArray)

				r.commandBuffer = append(r.commandBuffer, command{
					command:    serializedCommandArray,
					rawCommand: clientMsg.Msg,
					conn:       clientMsg.Conn,
				})
			}

			listOfCommands := [][]string{}
			for _, command := range r.commandBuffer {
				listOfCommands = append(listOfCommands, command.command)
			}
			fmt.Println("List of commands: ", listOfCommands)

		case closeConn := <-r.closeChan:
			slog.Info("Closing connection", "info", closeConn.RemoteAddr().String())
			closeConn.Close()
		}
	}
}

func (r *Replica) write(response string, conn net.Conn) {
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error encountered when writing response: %s", err.Error())
	}
}

/*
* Responsible for accepting new connections and appending the connection to the clients map
 */
func (r *Replica) acceptLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("error accepting connection", "err", err)
			continue
		}

		client := client.NewClient(conn, r.msgChan, r.closeChan)

		// Start reading messages from client
		go client.ReadLoop()
	}
}
