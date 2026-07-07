package database

import (
	"testing"

	"my-redis/interfaces/redis"
)

func TestHashHMSetHMGetAndDelete(t *testing.T) {
	db := makeDB()

	assertHashReply(t, db.Exec(nil, hashCmdLine("hmset", "user:1", "name", "alice", "age", "30")), "+OK\r\n")
	assertHashReply(t, db.Exec(nil, hashCmdLine("hmget", "user:1", "name", "age", "missing")), "*3\r\n$5\r\nalice\r\n$2\r\n30\r\n$-1\r\n")
	assertHashReply(t, db.Exec(nil, hashCmdLine("hstrlen", "user:1", "name")), ":5\r\n")
	assertHashReply(t, db.Exec(nil, hashCmdLine("hdel", "user:1", "age", "missing")), ":1\r\n")
	assertHashReply(t, db.Exec(nil, hashCmdLine("hexists", "user:1", "age")), ":0\r\n")
}

func TestHashHGetRejectsExtraArguments(t *testing.T) {
	db := makeDB()

	assertHashReply(t, db.Exec(nil, hashCmdLine("hset", "user:1", "name", "alice")), ":1\r\n")
	assertHashReply(t, db.Exec(nil, hashCmdLine("hget", "user:1", "name", "extra")), "-ERR wrong number of arguments for 'hget' command\r\n")
}

func hashCmdLine(args ...string) CmdLine {
	result := make(CmdLine, len(args))
	for i, arg := range args {
		result[i] = []byte(arg)
	}
	return result
}

func assertHashReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := string(reply.ToBytes()); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}
