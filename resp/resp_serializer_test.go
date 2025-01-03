package resp

import (
	"fmt"
	"testing"
)

func TestDeserializeRespBulk(t *testing.T) {
	// raw := "$4\r\nECHO\r\n"

	// desrializedCommand, err := RESPDeserializeCommand(raw)
	// if err != nil {
	// 	t.Errorf("Encountered error: %s", err.Error())
	// }

	// if desrializedCommand[0] != "ECHO" {
	// 	t.Errorf("Deserialize result RESP Bulk was incorrect. Received: %s", desrializedCommand)
	// }
}

func TestDeserializeRespArray(t *testing.T) {
	// raw := "*2\r\n$4\r\nECHO\r\n$9\r\nblueberry\r\n"
	raw := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n*3\r\n$3\r\nSET\r\n$3\r\nbar\r\n$3\r\n456\r\n*3\r\n$3\r\nSET\r\n$3\r\nbaz\r\n$3\r\n789\r\n"
	raw = "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n"
	raw = "+FULLRESYNC 75cd7bc10c49047e0d163660f3b90625b1af31dc 0\r\n$88\r\nREDIS0011\xfa\tredis-ver\x057.2.0\xfa\nredis-bits\xc0@\xfa\x05ctime\xc2m\b\xbce\xfa\bused-mem°\xc4\x10\x00\xfa\baof-base\xc0\x00\xff\xf0n;\xfe\xc0\xffZ\xa2*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n*3\r\n$3\r\nSET\r\n$3\r\nbar\r\n$3\r\n456\r\n*3\r\n$3\r\nSET\r\n$3\r\nbaz\r\n$3\r\n789\r\n"

	fmt.Println("ABOUT TO ")
	desrializedCommand, err := RESPDeserializeCommand(raw)
	fmt.Println("OUTPUT: ")
	fmt.Println(desrializedCommand)
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}

	// if desrializedCommand[0] != "ECHO" {
	// 	t.Errorf("Deserialize result RESP array was incorrect. Received: %s", desrializedCommand)
	//
	// }
	//
	// if desrializedCommand[1] != "blueberry" {
	// 	t.Errorf("Deserialize result RESP array was incorrect. Received: %s", desrializedCommand)
	//
	// }
}
