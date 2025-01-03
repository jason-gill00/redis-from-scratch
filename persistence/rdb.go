package persistence

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const (
	rdbOpCodeMetaData      = 250 // FA - metadata section
	rdbOpCodeDbSubsection  = 254 // FE - start of db section
	rdbOpCodeExpiryMs      = 252 // FC - indicates key has exp in ms
	rdbOpCodeHashTableSize = 251 // FB - indivicates start of hash table size
	rdbOpCodeEnd           = 255 // FF - end of rdb file
)

type data struct {
	Value      string
	Expiration *uint64
}

type database = map[string]data

type RdbFile struct {
	header   string
	metadata map[string]string
	Database database
}

func ParseRdb(data []byte) (*RdbFile, error) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read header
	header, err := readHeader(reader)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(header), "REDIS") {
		return nil, fmt.Errorf("invalid rdb header: %s", string(header))
	}

	// metadata := make(map[string]string)
	db := make(database)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case rdbOpCodeMetaData:
			// TODO: read metadata
			continue
		case rdbOpCodeDbSubsection:
			db, err = readDatabase(reader)
			if err != nil {
				return nil, err
			}
		case rdbOpCodeEnd:
			return &RdbFile{header: string(header), metadata: nil, Database: db}, nil
		default:
			continue
		}

	}
}

func ParseRdbFile(rdbFile string) (*RdbFile, error) {
	rdb, err := os.Open(rdbFile)
	if err != nil {
		return nil, err
	}
	defer rdb.Close()
	reader := bufio.NewReader(rdb)

	// Read header
	header, err := readHeader(reader)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(header), "REDIS") {
		return nil, fmt.Errorf("invalid rdb header: %s", string(header))
	}

	// metadata := make(map[string]string)
	db := make(database)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case rdbOpCodeMetaData:
			// TODO: read metadata
			continue
		case rdbOpCodeDbSubsection:
			db, err = readDatabase(reader)
			if err != nil {
				return nil, err
			}
		case rdbOpCodeEnd:
			return &RdbFile{header: string(header), metadata: nil, Database: db}, nil
		default:
			continue
		}

	}
}

func readDatabase(reader *bufio.Reader) (map[string]data, error) {
	_, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	db := database{}

	for {
		// Read opcode without moving the reader
		peekOpcode, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}
		if peekOpcode[0] == rdbOpCodeEnd {
			return db, nil
		}
		// Read opcode
		opcode, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		switch opcode {
		case rdbOpCodeEnd:
			return db, nil
		case rdbOpCodeHashTableSize:
			_, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			_, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}
		case rdbOpCodeExpiryMs:
			// Next 8 bytes specify the timestamp
			timestampBuf := make([]byte, 8)
			t, err := reader.Read(timestampBuf)
			// Convert the byte slice to an int64 (litt-endian encoding)
			timestamp := binary.LittleEndian.Uint64(timestampBuf[:t])

			if err != nil {
				return nil, err
			}

			_, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}
			key, value, err := readKeyValue(reader)
			if err != nil {
				return nil, err
			}
			db[string(key)] = data{Value: string(value), Expiration: &timestamp}
		default:
			// No expiration is specified so we just need to read key/value
			key, value, err := readKeyValue(reader)
			if err != nil {
				return nil, err
			}
			db[string(key)] = data{Value: string(value), Expiration: nil}

		}
	}
}

func readKeyValue(reader *bufio.Reader) (string, string, error) {
	keyLen, err := reader.ReadByte()
	if err != nil {
		return "", "", err
	}
	keyBuf := make([]byte, keyLen)
	k, err := reader.Read(keyBuf)
	if err != nil {
		return "", "", err
	}
	valueLen, err := reader.ReadByte()
	if err != nil {
		return "", "", err
	}
	valueBuf := make([]byte, valueLen)
	v, err := reader.Read(valueBuf)
	if err != nil {
		return "", "", err
	}

	return string(keyBuf[:k]), string(valueBuf[:v]), nil
}

/*
* This fn reads the metadata from the rdb file TODO: currently this just supports a single key/value pair.
* extend this to support n key/values
 */
func readMetadata(reader *bufio.Reader) (map[string]string, error) {
	attrByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	attrBuffer := make([]byte, attrByte)
	n, err := reader.Read(attrBuffer)
	if err != nil {
		return nil, err
	}
	attrName := attrBuffer[:n]

	attrValByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	attrValBuffer := make([]byte, attrValByte)
	m, err := reader.Read(attrValBuffer)
	if err != nil {
		return nil, err
	}

	return map[string]string{string(attrName): string(attrValBuffer[:m])}, nil
}

func readHeader(reader *bufio.Reader) ([]byte, error) {
	// Create a buffer for the header (size of the header is 9 bytes)
	buffer := make([]byte, 9)
	n, err := reader.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:n], nil
}
