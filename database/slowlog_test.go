package database

import (
	"testing"
	"time"

	"my-redis/interfaces/redis"
)

func TestSlowLoggerRecordsLatestEntriesFirst(t *testing.T) {
	logger := NewSlowLogger(2, 0)

	logger.Record(time.Now().Add(-3*time.Millisecond), slowlogCmdLine("set", "a", "1"), "client-1")
	logger.Record(time.Now().Add(-2*time.Millisecond), slowlogCmdLine("set", "b", "2"), "client-2")
	logger.Record(time.Now().Add(-1*time.Millisecond), slowlogCmdLine("get", "b"), "client-3")

	if got := logger.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
	entries := logger.GetEntries(10)
	if len(entries) != 2 {
		t.Fatalf("GetEntries len = %d, want 2", len(entries))
	}
	if entries[0].ID != 3 || entries[1].ID != 2 {
		t.Fatalf("entry ids = %d,%d, want 3,2", entries[0].ID, entries[1].ID)
	}
}

func TestSlowLoggerHandleCommands(t *testing.T) {
	logger := NewSlowLogger(4, 0)
	logger.Record(time.Now().Add(-time.Millisecond), slowlogCmdLine("ping"), "client")

	assertSlowlogReply(t, logger.HandleSlowlogCommand(slowlogCmdLine("slowlog", "len")), ":1\r\n")
	assertSlowlogReply(t, logger.HandleSlowlogCommand(slowlogCmdLine("slowlog", "reset")), "+OK\r\n")
	assertSlowlogReply(t, logger.HandleSlowlogCommand(slowlogCmdLine("slowlog", "len")), ":0\r\n")
}

func slowlogCmdLine(args ...string) CmdLine {
	line := make(CmdLine, len(args))
	for i, arg := range args {
		line[i] = []byte(arg)
	}
	return line
}

func assertSlowlogReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := string(reply.ToBytes()); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}
