package database

import (
	"testing"

	"my-redis/interfaces/redis"
)

func TestSetStoreCommands(t *testing.T) {
	db := makeDB()

	assertSetReply(t, db.Exec(nil, setCmdLine("sadd", "a", "1", "2")), ":2\r\n")
	assertSetReply(t, db.Exec(nil, setCmdLine("sadd", "b", "2", "3")), ":2\r\n")

	assertSetReply(t, db.Exec(nil, setCmdLine("sinterstore", "inter", "a", "b")), ":1\r\n")
	assertSetReply(t, db.Exec(nil, setCmdLine("scard", "inter")), ":1\r\n")
	assertSetReply(t, db.Exec(nil, setCmdLine("sismember", "inter", "2")), ":1\r\n")

	assertSetReply(t, db.Exec(nil, setCmdLine("sunionstore", "union", "a", "b")), ":3\r\n")
	assertSetReply(t, db.Exec(nil, setCmdLine("scard", "union")), ":3\r\n")

	assertSetReply(t, db.Exec(nil, setCmdLine("sdiffstore", "diff", "a", "b")), ":1\r\n")
	assertSetReply(t, db.Exec(nil, setCmdLine("sismember", "diff", "1")), ":1\r\n")
}

func TestSetPopRemovesMembers(t *testing.T) {
	db := makeDB()

	assertSetReply(t, db.Exec(nil, setCmdLine("sadd", "letters", "a", "b")), ":2\r\n")
	reply := db.Exec(nil, setCmdLine("spop", "letters", "1"))
	if got := setReplyString(reply); got != "*1\r\n$1\r\na\r\n" && got != "*1\r\n$1\r\nb\r\n" {
		t.Fatalf("spop reply = %q, want one popped member", got)
	}
	assertSetReply(t, db.Exec(nil, setCmdLine("scard", "letters")), ":1\r\n")
}

func setCmdLine(args ...string) CmdLine {
	result := make(CmdLine, len(args))
	for i, arg := range args {
		result[i] = []byte(arg)
	}
	return result
}

func assertSetReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := setReplyString(reply); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}

func setReplyString(reply redis.Reply) string {
	return string(reply.ToBytes())
}
