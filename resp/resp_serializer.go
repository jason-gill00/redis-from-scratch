package resp

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// RESP Types
const (
	RESPArray  = "*"
	RESPBulk   = "$"
	RESPSimple = "+"
)

// Hardcoded serialized responses
const (
	RESPNil = "$-1\r\n"
)

// func RESPDeserializeCommand(rawCommand string) ([]string, error) {
// 	reader := bufio.NewReader(strings.NewReader(rawCommand))
// 	parsedResp, err := parseResp(reader)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return parsedResp, nil
// }

func RESPDeserializeCommand(rawCommand string) ([][]string, error) {
	reader := bufio.NewReader(strings.NewReader(rawCommand))
	var allParsedResp [][]string

	for {
		parsedResp, err := parseResp(reader)
		if err != nil {
			// Break if end of input
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		allParsedResp = append(allParsedResp, parsedResp)
	}

	return allParsedResp, nil
}

func RESPSerializeRESPArray(elements []string) string {
	str := fmt.Sprintf("*%d\r\n", len(elements))

	for _, elem := range elements {
		str += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
	}

	return str
}

func RESPSerializeSimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func RESPSerializeBulkString(str string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
}

func RESPSerializeFile(str string) string {
	return fmt.Sprintf("$%d\r\n%s", len(str), str)
}

func parseResp(reader *bufio.Reader) ([]string, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch string(prefix) {
	case RESPBulk:
		return handleRespBulk(reader)
	case RESPArray:
		return handleRespArray(reader)
	case RESPSimple:
		return handleRespSimple(reader)
	default:
		fmt.Println("Unexpected RESP TYPE", prefix)
		return nil, nil
	}
}

func handleRespSimple(reader *bufio.Reader) ([]string, error) {
	str, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return []string{strings.TrimSuffix(str, "\r\n")}, nil
}

func handleRespArray(reader *bufio.Reader) ([]string, error) {
	strLength, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(strings.TrimSuffix(strLength, "\r\n"))

	arr := make([]string, length)
	for i := 0; i < length; i++ {
		command, err := parseResp(reader)
		if err != nil {
			return nil, err
		}

		arr[i] = command[0]

	}
	return arr, nil
}

func handleRespBulk(reader *bufio.Reader) ([]string, error) {
	// The first character is the length of the string
	strLength, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(strings.TrimSuffix(strLength, "\r\n"))
	if err != nil {
		return nil, err
	}

	// Read x bytes from the buffer
	data := make([]byte, length)
	n, err := reader.Read(data)
	if err != nil {
		return nil, err
	}

	// Check if there is an extra \r\n and skip it
	// The `ReadString` might return an extra \r\n for the bulk string terminator.
	if extra, _ := reader.Peek(2); string(extra) == "\r\n" {
		_, _ = reader.ReadString('\n') // Discard the \r\n terminator
	}

	return []string{string(data[:n])}, nil
}
