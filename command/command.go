package command

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jason-gill00/redis-from-scratch/persistence"
	"github.com/jason-gill00/redis-from-scratch/resp"
)

const (
	GET      = "GET"
	SET      = "SET"
	PING     = "PING"
	ECHO     = "ECHO"
	CONFIG   = "CONFIG"
	KEY      = "KEYS"
	INFO     = "INFO"
	REPLCONF = "REPLCONF"
	PSYNC    = "PSYNC"
)

/*
* Takes in deserialized command array (["SET", "KEY", "VALUE"])
* and returns response that the connected client would expect
 */
func CacheCommandHandler(command []string, store *persistence.Store, config map[string]string, replicationConfig map[string]string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("no command found")
	}

	switch strings.ToUpper(command[0]) {
	case PING:
		return pingCommandHandler(), nil
	case ECHO:
		return echoCommandHandler(command), nil
	case GET:
		return getCommandHandler(command, store), nil
	case SET:
		return setCommandHandler(command, store), nil
	case CONFIG:
		return configCommandHandler(command, config)
	case KEY:
		return keyCommandHandler(command, config)
	case INFO:
		return infoCommandHandler(command, replicationConfig)
	case REPLCONF:
		return replConfCommandHandler(command, replicationConfig), nil
	case PSYNC:
		return psyncCommandHandler(command), nil
	default:
		return "", nil
	}
}

func psyncCommandHandler(command []string) string {
	// TODO: implement replication id
	replicationId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	fmt.Println("PSYNC command received")
	if command[1] == "?" && command[2] == "-1" {
		fmt.Println("PSYNC command received returning fullresync")
		return resp.RESPSerializeSimpleString(fmt.Sprintf("FULLRESYNC %s 0", replicationId))
	}

	fmt.Println("returning nothing")

	return ""
}

func replConfCommandHandler(command []string, replicationConfig map[string]string) string {
	fmt.Println("REPLCONF command received: ", command[1])
	fmt.Println("REPLCONF command received: ", command)
	if strings.ToLower(command[1]) == "getack" {
		offset := replicationConfig["slave_repl_offset"]
		fmt.Println("REPLCONF GETACK command received: ", offset)
		return resp.RESPSerializeRESPArray([]string{"REPLCONF", "ACK", offset})
	}

	return resp.RESPSerializeSimpleString("OK")
}

func infoCommandHandler(command []string, replicationConfig map[string]string) (string, error) {
	fmt.Println("INFO command received: ", command)
	fmt.Println(command)
	if len(command) == 1 {
		return "", fmt.Errorf("no command found")
	}
	if command[1] != "replication" {
		return "", fmt.Errorf("invalid info command: %s", command[1])
	}

	// TODO: Implement replication id and offset
	replicationId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	offset := "0"

	if replicationConfig["replicaof"] != "" {
		return resp.RESPSerializeBulkString("role:slave"), nil
	}

	return resp.RESPSerializeBulkString(fmt.Sprintf("role:master\nmaster_replid:%s\nmaster_repl_offset:%s", replicationId, offset)), nil
}

func keyCommandHandler(command []string, config map[string]string) (string, error) {
	if command[1] != "*" {
		return "", fmt.Errorf("invalid key command: %s", command[1])

	}

	parsedRdb, err := persistence.ParseRdbFile(config["dir"] + "/" + config["dbFileName"])
	if err != nil {
		return "", err
	}

	keys := []string{}
	for key, _ := range parsedRdb.Database {
		keys = append(keys, key)
	}

	return resp.RESPSerializeRESPArray(keys), nil

}

func configCommandHandler(command []string, config map[string]string) (string, error) {
	if command[1] != GET {
		return "", fmt.Errorf("invalid config command: %s", command[1])
	}
	configParam := command[2]

	if val, ok := config[configParam]; ok {
		return resp.RESPSerializeRESPArray([]string{configParam, val}), nil
	}

	return resp.RESPNil, nil
}

func pingCommandHandler() string {
	// return RESPSerializeSimpleString("PONG")
	return resp.RESPSerializeRESPArray([]string{"PONG"})
}

func echoCommandHandler(command []string) string {
	return resp.RESPSerializeSimpleString(command[1])
}

func setCommandHandler(command []string, store *persistence.Store) string {
	key, value := command[1], command[2]
	expiration := getExpiration(command)

	store.Set(key, []byte(value), expiration)

	return resp.RESPSerializeSimpleString("OK")

}

func getCommandHandler(command []string, store *persistence.Store) string {
	key := command[1]

	if val, ok := store.Get(key); ok {
		return resp.RESPSerializeSimpleString(string(val))
	}

	slog.Info("No key found with key", "info", command)
	return resp.RESPNil
}

// getExpiration checks for an expiration value in the command array.
// Returns the expiration time and a boolean indicating if an expiration was found.
func getExpiration(command []string) *time.Time {
	for i, cmd := range command {
		if strings.ToUpper(cmd) == "PX" {
			// Ensure there's a value after "PX"
			if i+1 >= len(command) {
				slog.Info("Encountered PX but no expiration value", "info", command)
				return nil
			}

			// Convert the next argument to an integer (milliseconds)
			exp, err := strconv.Atoi(command[i+1])
			if err != nil {
				slog.Info("Invalid expiration value:", "info", command[i+1])
				return nil
			}

			// Convert milliseconds to duration and calculate future time
			duration := time.Duration(exp) * time.Millisecond
			expiration := time.Now().Add(duration)
			return &expiration
		}
	}

	// No expiration found
	return nil
}
