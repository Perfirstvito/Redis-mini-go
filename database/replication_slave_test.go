package database

import (
	"sync/atomic"
	"testing"

	"my-redis/redis/parser"
	"my-redis/redis/protocol"
)

func TestReplicationSlaveParseFullResyncHandshake(t *testing.T) {
	server := &Server{
		slaveStatus: initReplSlaveStatus(),
	}
	ch := make(chan *parser.Payload, 1)
	ch <- &parser.Payload{Data: protocol.MakeStatusReply("FULLRESYNC repl-id 42")}
	server.slaveStatus.masterChan = ch

	full, err := server.parsePsyncHandshake()
	if err != nil {
		t.Fatalf("parsePsyncHandshake() error = %v", err)
	}
	if !full || server.slaveStatus.replId != "repl-id" || server.slaveStatus.replOffset != 42 {
		t.Fatalf("handshake state = full %v id %q offset %d", full, server.slaveStatus.replId, server.slaveStatus.replOffset)
	}
}

func TestReplicationSlaveOfNoneResetsState(t *testing.T) {
	server := &Server{
		slaveStatus: initReplSlaveStatus(),
		role:        slaveRole,
	}
	server.slaveStatus.masterHost = "127.0.0.1"
	server.slaveStatus.masterPort = 6379
	server.slaveStatus.replId = "repl-id"
	server.slaveStatus.replOffset = 12

	server.slaveOfNone()

	if server.slaveStatus.masterHost != "" || server.slaveStatus.masterPort != 0 {
		t.Fatalf("master address = %q:%d, want empty", server.slaveStatus.masterHost, server.slaveStatus.masterPort)
	}
	if server.slaveStatus.replId != "" || server.slaveStatus.replOffset != -1 {
		t.Fatalf("repl state = %q/%d, want empty/-1", server.slaveStatus.replId, server.slaveStatus.replOffset)
	}
	if got := atomic.LoadInt32(&server.role); got != masterRole {
		t.Fatalf("role = %d, want masterRole", got)
	}
}
