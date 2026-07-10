package database

import (
	"testing"

	"my-redis/interfaces/redis"
)

func TestSortedSetRangeAndScoreCommands(t *testing.T) {
	db := makeDB()

	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zadd", "scores", "1", "alice", "2.5", "bob", "3", "carol")), ":3\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zrange", "scores", "0", "-1", "withscores")), "*6\r\n$5\r\nalice\r\n$1\r\n1\r\n$3\r\nbob\r\n$3\r\n2.5\r\n$5\r\ncarol\r\n$1\r\n3\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zcount", "scores", "1", "2.5")), ":2\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zincrby", "scores", "3", "alice")), "$1\r\n4\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zrevrank", "scores", "alice")), ":0\r\n")
}

func TestSortedSetRemoveRangeCommands(t *testing.T) {
	db := makeDB()

	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zadd", "scores", "1", "a", "2", "b", "3", "c")), ":3\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zremrangebyscore", "scores", "1", "2")), ":2\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zscore", "scores", "a")), "$-1\r\n")
	assertSortedSetReply(t, db.Exec(nil, sortedSetCmdLine("zcard", "scores")), ":1\r\n")
}

func sortedSetCmdLine(args ...string) CmdLine {
	result := make(CmdLine, len(args))
	for i, arg := range args {
		result[i] = []byte(arg)
	}
	return result
}

func assertSortedSetReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := string(reply.ToBytes()); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}
