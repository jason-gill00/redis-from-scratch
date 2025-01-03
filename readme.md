## Project

In this project I create a redis server from scratch. This server is responsible for accepting redis commands from clients, master slave replication, RESP serialization/deserialization, and RDB persistence.

## RESP Serialization/Deserialization

Let’s start by discussing RESP (Redis serialization protocol). This protocal is used in the communication between a Redis server and conencted clients (over a TCP connection). The redis server accepts commands that are sereialized using the RESP protocal, deserializes/proccesses the command and returns a response serialized using RESP.

RESP data can be split into three types: simple, bulk or aggregate. `\r\n` (CRLF) can be used to denote the seperate parts of a comand.

| **RESP data type** | **Category** | **First byte** | Format | Example |
| --- | --- | --- | --- | --- |
| [Simple strings](https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings) | Simple | **`+`** | `+OK\r\n` | `+OK\r\n` |
| [Simple Errors](https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors) | Simple | **`-`** | `-Error message \r\n` | `-Error message \r\n` |
| [Integers](https://redis.io/docs/latest/develop/reference/protocol-spec/#integers) | Simple | **`:`** | `:[<+|→|<value>\r\n` | `:0\r\n` |
| [Bulk strings](https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings) | Aggregate | **`$`** | `$<length>\r\n<data\r\n` | `$5\r\nhello\r\n` |
| [Arrays](https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays) | Aggregate | **`*`** | `$-1\r\n` | `$-1\r\n` |
| [Nulls](https://redis.io/docs/latest/develop/reference/protocol-spec/#nulls) | Simple | **`_`** | `*<number-of-elements>\r\n<element-1>...<element-n>` | `*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n` |

## Commands

These are the different commands this Redis server supports:

| Command | RESP Request | RESP Response | Description |
| --- | --- | --- | --- |
| PING | `*1\r\n$4\r\nPING\r\n` | `+PONG\r\n` | Check whether the server is healthy |
| SET | `*5\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n$2\r\nEX\r\n$3\r\n100\r\n` | `+OK\r\n` | Set a key to a value (with an optional expiration) |
| GET | `*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n` | `$3\r\nBAR\r\n` or `$-1\r\n` | Retrieve the value of a key |
| INFO | `*2\r\n$4\r\nINFO\r\n$11\r\nreplication\r\n` | `$11\r\nrole:master\r\n`  | Returns information about the server |
| REPLCONF GETACK | `*2\r\n$7\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1*\r\n3` | `*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n` | Master sends this command to get the replica offsaet |
| PSYNC | `*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n` | `+FULLRESYNC <REPL_ID> 0\r\n` | Synchronize the state of the replica with the master |

## Master/Slave Replications

In this server I implemented master/slave replication. The master is responsible for receiving commands from the client. If a write command is received the master will replicate the command to all connected replicas.

Before the master can send commands to the replicas, the replica needs to preform a handshake with the master. The handshake is composed of three parts:

- Replica sends a `PING` to the master
- Replica sends a `REPLCONF` twict to the master
    - `REPLCONF listening-port <port>`
        - Notify the master of what port the replica is listening on
    - `REPLCONF CAPA psync2`
        - Replica notifys the master of it’s capabilities
- The replica sends `PSYNC` to the master
    - Used to synchronize the state of the replica with the master
    - The master will respond with FULLRESYNC and proceed to send the masters rdb file to the replica

The master and slaves use replication offsets to determine whether they are in sync with one another. The replicaation offset corresponds to how many bytes of commands have been added to the replication stream. The master periodically sends `REPLCONF GETACK` commands to the replicas to ensure the replicas are in sync.