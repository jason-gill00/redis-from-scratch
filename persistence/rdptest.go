package persistence

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func createTempRdbFile(t *testing.T, header string, metadata map[string]string, db database) string {
	t.Helper()

	var buffer bytes.Buffer

	// Write the header
	buffer.WriteString(header)

	// Write metadata
	for key, value := range metadata {
		buffer.WriteByte(byte(rdbOpCodeMetaData)) // Metadata opcode (FA)
		buffer.WriteByte(byte(len(key)))          // Length of key
		buffer.WriteString(key)                   // Key
		buffer.WriteByte(byte(len(value)))        // Length of value
		buffer.WriteString(value)                 // Value
	}

	// Start database subsection
	buffer.WriteByte(byte(rdbOpCodeDbSubsection)) // Start of DB subsection (FE)
	buffer.WriteByte(0x00)                        // Database index (size-encoded, here it's 0)

	// Write hash table size information
	buffer.WriteByte(byte(rdbOpCodeHashTableSize)) // Hash table size opcode (FB)
	buffer.WriteByte(byte(len(db)))                // Total hash table size
	expiryCount := 0
	for _, data := range db {
		if data.Expiration != nil {
			expiryCount++
		}
	}
	buffer.WriteByte(byte(expiryCount)) // Number of keys with expirations

	// Write key-value pairs
	for key, data := range db {
		if data.Expiration != nil {
			fmt.Println("data.expiration", data.Expiration)
			buffer.WriteByte(byte(rdbOpCodeExpiryMs)) // Expiry opcode (FC)
			// buffer.Write(*data.expiration)            // Expiry timestamp (8 bytes)
		}
		buffer.WriteByte(0x00)                  // Value type (0 = string)
		buffer.WriteByte(byte(len(key)))        // Key length
		buffer.WriteString(key)                 // Key
		buffer.WriteByte(byte(len(data.Value))) // Value length
		buffer.WriteString(data.Value)          // Value

	}

	// Create temp file
	file, err := os.CreateTemp("", "test_rdb_*.rdb")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		t.Fatalf("Failed to write RDB content: %v", err)
	}

	file.Close()
	return file.Name()
}

// func TestParseRdbFile(t *testing.T) {
// 	content := "REDIS0006"
// 	metadata := map[string]string{
// 		"redis-ver": "6.0.16",
// 		"key1":      "value1",
// 	}
//
// 	// Define a mock database with a key-value pair and expiration
// 	db := database{
// 		"key1": {
// 			value:      "value1",
// 			expiration: &[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, // Timestamp in ms
// 		},
// 		"key2": {
// 			value:      "value2",
// 			expiration: nil, // No expiration
// 		},
// 	}
//
// 	tempFile := createTempRdbFile(t, content, metadata, db)
// 	defer os.Remove(tempFile)
//
// 	_, err := ParseRdbFile(tempFile)
// 	if err != nil {
// 		fmt.Printf("Encountered error parsing rdb: %s \n", err.Error())
// 		return
// 	}
// }

func TestParseRdbFileWithDatabaseSection(t *testing.T) {
	header := "REDIS0011"
	metadata := map[string]string{
		"redis-ver":  "7.2.0",
		"redis-bits": "@",
	}
	db := database{
		"orange": data{
			Value:      "strawberry",
			Expiration: nil,
		},
	}

	// Create temp RDB file
	tempFile := createTempRdbFile(t, header, metadata, db)
	defer os.Remove(tempFile)

	// Parse RDB file
	parsedRdb, err := ParseRdbFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to parse RDB file: %v", err)
	}

	fmt.Println("DONE")
	fmt.Println(parsedRdb)
}

func TestParseRdbFile(t *testing.T) {
	// Define header and metadata
	header := "REDIS0011"
	metadata := map[string]string{
		"redis-ver": "6.0.16",
	}

	// Define database with keys and expirations
	db := database{
		"foobar": {
			Value:      "bazqux",
			Expiration: &[]byte{0x15, 0x72, 0xE7, 0x07, 0x8F, 0x01, 0x00, 0x00}, // Expiry timestamp
		},
		"foo": {
			Value:      "bar",
			Expiration: nil, // No expiration
		},
	}

	// Create the temp RDB file
	rdbFile := createTempRdbFile(t, header, metadata, db)

	// Parse the RDB file
	rdb, err := ParseRdbFile(rdbFile)
	if err != nil {
		t.Fatalf("Failed to parse RDB file: %v", err)
	}

	fmt.Println("DONE")
	fmt.Println(rdb)
}

func TestCreateHexFile(t *testing.T) {
	// Hex values from the given example
	hexData := []byte{
		0x52, 0x45, 0x44, 0x49, 0x53, 0x30, 0x30, 0x31, 0x31, 0xfa, 0x09, 0x72, 0x65, 0x64, 0x69, 0x73,
		0x2d, 0x76, 0x65, 0x72, 0x05, 0x37, 0x2e, 0x32, 0x2e, 0x30, 0xfa, 0x0a, 0x72, 0x65, 0x64, 0x69,
		0x73, 0x2d, 0x62, 0x69, 0x74, 0x73, 0xc0, 0x40, 0xfe, 0x00, 0xfb, 0x05, 0x05, 0xfc, 0x00, 0x0c,
		0x28, 0x8a, 0xc7, 0x01, 0x00, 0x00, 0x00, 0x06, 0x6f, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x06, 0x62,
		0x61, 0x6e, 0x61, 0x6e, 0x61, 0xfc, 0x00, 0x0c, 0x28, 0x8a, 0xc7, 0x01, 0x00, 0x00, 0x00, 0x09,
		0x72, 0x61, 0x73, 0x70, 0x62, 0x65, 0x72, 0x72, 0x79, 0x09, 0x70, 0x69, 0x6e, 0x65, 0x61, 0x70,
		0x70, 0x6c, 0x65, 0xfc, 0x00, 0x0c, 0x28, 0x8a, 0xc7, 0x01, 0x00, 0x00, 0x00, 0x05, 0x61, 0x70,
		0x70, 0x6c, 0x65, 0x05, 0x6d, 0x61, 0x6e, 0x67, 0x6f, 0xfc, 0x00, 0x9c, 0xef, 0x12, 0x7e, 0x01,
		0x00, 0x00, 0x00, 0x04, 0x70, 0x65, 0x61, 0x72, 0x09, 0x72, 0x61, 0x73, 0x70, 0x62, 0x65, 0x72,
		0x72, 0x79, 0xfc, 0x00, 0x0c, 0x28, 0x8a, 0xc7, 0x01, 0x00, 0x00, 0x00, 0x05, 0x67, 0x72, 0x61,
		0x70, 0x65, 0x05, 0x61, 0x70, 0x70, 0x6c, 0x65, 0xff, 0xa7, 0x2c, 0x56, 0x50, 0x50, 0x30, 0x3c,
		0x67, 0x0a,
	} // Create a temporary file

	tempFile, err := os.CreateTemp("", "test_hex_file.rdb")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the file afterward

	// Write the hex data to the file
	if _, err := tempFile.Write(hexData); err != nil {
		t.Fatalf("Failed to write hex data to file: %v", err)
	}

	tempFile.Close()

	defer os.Remove(tempFile.Name())

	// Parse RDB file
	parsedRdb, err := ParseRdbFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse RDB file: %v", err)
	}

	fmt.Println("DONE")
	fmt.Println(parsedRdb)

}
