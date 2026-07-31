package database

import (
	"strings"
	"testing"

	"my-redis/redis/connection"
)

func TestReplicationBacklogSnapshotAndOffset(t *testing.T) {
	backlog := &replBacklog{beginOffset: 10, currentOffset: 10}
	backlog.appendBytes([]byte("abcdef"))

	if !backlog.isValidOffset(10) || !backlog.isValidOffset(15) || backlog.isValidOffset(16) {
		t.Fatalf("unexpected backlog offset validity: begin=%d current=%d", backlog.beginOffset, backlog.currentOffset)
	}
	if got, current := backlog.getSnapshotAfter(13); string(got) != "def" || current != 16 {
		t.Fatalf("snapshot after offset = %q,%d, want def,16", string(got), current)
	}
}

func TestReplicationMasterPartialSyncWritesBacklog(t *testing.T) {
	server := &Server{}
	server.initMasterStatus()
	server.masterStatus.replId = "repl-id"
	server.masterStatus.backlog.appendBytes([]byte("abcdef"))

	conn := connection.NewFakeConn()
	slave := &slaveClient{conn: conn}
	err := server.masterTryPartialSyncWithSlave(slave, "repl-id", 3)
	if err != nil {
		t.Fatalf("masterTryPartialSyncWithSlave() error = %v", err)
	}
	if got := string(conn.Bytes()); !strings.HasPrefix(got, "+CONTINUE repl-id\r\n") || !strings.HasSuffix(got, "def") {
		t.Fatalf("partial sync bytes = %q", got)
	}
	if slave.state != slaveStateOnline || slave.offset != 6 {
		t.Fatalf("slave state/offset = %d/%d, want online/6", slave.state, slave.offset)
	}
}

func TestReplicationAofListenerAppendsCommandsToBacklog(t *testing.T) {
	server := &Server{}
	server.initMasterStatus()
	listener := &replAofListener{
		mdb:     server,
		backlog: server.masterStatus.backlog,
	}

	listener.Callback([]CmdLine{replicationCmdLine("set", "k", "v")})
	got, current := server.masterStatus.backlog.getSnapshot()
	if current != int64(len(got)) || len(got) == 0 {
		t.Fatalf("backlog snapshot len/current = %d/%d", len(got), current)
	}
	if !strings.Contains(string(got), "set") || !strings.Contains(string(got), "k") || !strings.Contains(string(got), "v") {
		t.Fatalf("backlog bytes = %q, want encoded set command", string(got))
	}
}

func replicationCmdLine(args ...string) CmdLine {
	line := make(CmdLine, len(args))
	for i, arg := range args {
		line[i] = []byte(arg)
	}
	return line
}
