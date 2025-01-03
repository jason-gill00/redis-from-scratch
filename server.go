package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jason-gill00/redis-from-scratch/master"
	pers "github.com/jason-gill00/redis-from-scratch/persistence"
	"github.com/jason-gill00/redis-from-scratch/replica"
)

var _ = net.Listen
var _ = os.Exit

var dir = flag.String("dir", "", "RDB file store directory")
var dbFileName = flag.String("dbfilename", "dump.rdb", "RDB dump")
var port = flag.String("port", "6379", "Port to listen on")
var replicaOf = flag.String("replicaof", "", "Replicate to another redis server")

func readRdbFile(dir string, dbFileName string, store *pers.Store) {
	parsedRdb, err := pers.ParseRdbFile(dir + "/" + dbFileName)
	if err != nil {
		// If file does not exist, just continue
		if os.IsNotExist(err) {
			slog.Info("RDB file does not exist, continuing without loading data")
			return
		}

		fmt.Printf("Encountered error parsing rdb: %s \n", err.Error())
		return
	}
	for key, value := range parsedRdb.Database {

		if value.Expiration == nil {
			store.Set(key, []byte(value.Value), nil)
			continue
		} else {
			expiration := time.Unix(int64(*value.Expiration)/1000, 0).UTC()
			store.Set(key, []byte(value.Value), &expiration)
		}
	}
}

func main() {
	flag.Parse()

	config := map[string]string{
		"dir":        *dir,
		"dbFileName": *dbFileName,
	}

	store := pers.NewStore()

	// If a rdb file is provided, read the database and store it in the server
	if config["dir"] != "" && config["dbFileName"] != "" {
		readRdbFile(config["dir"], config["dbFileName"], store)
	}

	replicationConfig := map[string]string{
		"replicaof": *replicaOf,
	}

	// If the server is a replica, start the replica server
	if replicationConfig["replicaof"] != "" {
		replicationConfig["slave_repl_offset"] = "0"
		replica := replica.NewReplica(replicationConfig, store, config, *port)
		replica.Start()
		return
	}

	// If the server is a master, start the master server
	master := master.NewMaster(replicationConfig, store, config, *port)
	master.Start()
}
