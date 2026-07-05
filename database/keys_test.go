package database

import (
	"sync/atomic"
	"testing"
	"time"

	idatabase "my-redis/interfaces/database"
	"my-redis/interfaces/redis"
	"my-redis/redis/connection"
)

func TestKeysRenamePreservesValueAndTTL(t *testing.T) {
	db := makeDB()

	db.PutEntity("old", &idatabase.DataEntity{Data: []byte("value")})
	db.Expire("old", time.Now().Add(60*time.Second))
	t.Cleanup(func() {
		db.Persist("new")
	})

	assertKeyReply(t, db.Exec(nil, keyCmdLine("rename", "old", "new")), "+OK\r\n")
	assertKeyReply(t, db.Exec(nil, keyCmdLine("exists", "old")), ":0\r\n")
	entity, ok := db.GetEntity("new")
	if !ok {
		t.Fatal("renamed key does not exist")
	}
	if got := string(entity.Data.([]byte)); got != "value" {
		t.Fatalf("renamed value = %q, want value", got)
	}
	if got := keyReplyString(db.Exec(nil, keyCmdLine("ttl", "new"))); got != ":59\r\n" && got != ":60\r\n" {
		t.Fatalf("ttl after rename = %q, want about 60 seconds", got)
	}
}

func TestKeysToTTLCmd(t *testing.T) {
	db := makeDB()

	if got, want := keyCmdLineToStrings(toTTLCmd(db, "key").Args), []string{"PERSIST", "key"}; !sameKeyStrings(got, want) {
		t.Fatalf("toTTLCmd without ttl = %v, want %v", got, want)
	}

	expireAt := time.UnixMilli(123456789)
	db.ttlMap.Put("key", expireAt)
	if got, want := keyCmdLineToStrings(toTTLCmd(db, "key").Args), []string{"PEXPIREAT", "key", "123456789"}; !sameKeyStrings(got, want) {
		t.Fatalf("toTTLCmd with ttl = %v, want %v", got, want)
	}
}

func TestKeysCopyBetweenDatabases(t *testing.T) {
	server := makeTestServerWithDBs(2)
	server.mustSelectDB(0).PutEntity("src", &idatabase.DataEntity{Data: []byte("value")})

	conn := connection.NewFakeConn()
	conn.SelectDB(0)
	assertKeyReply(t, execCopy(server, conn, keyCmdLine("src", "dst", "db", "1")), ":1\r\n")

	entity, ok := server.mustSelectDB(1).GetEntity("dst")
	if !ok {
		t.Fatal("COPY did not create destination key")
	}
	if got := string(entity.Data.([]byte)); got != "value" {
		t.Fatalf("copied value = %q, want value", got)
	}
}

func keyCmdLine(args ...string) CmdLine {
	result := make(CmdLine, len(args))
	for i, arg := range args {
		result[i] = []byte(arg)
	}
	return result
}

func assertKeyReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := keyReplyString(reply); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}

func keyReplyString(reply redis.Reply) string {
	return string(reply.ToBytes())
}

func keyCmdLineToStrings(line CmdLine) []string {
	result := make([]string, len(line))
	for i, arg := range line {
		result[i] = string(arg)
	}
	return result
}

func sameKeyStrings(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func makeTestServerWithDBs(dbCount int) *Server {
	server := &Server{
		dbSet: make([]*atomic.Value, dbCount),
	}
	for i := range server.dbSet {
		db := makeDB()
		db.index = i
		holder := &atomic.Value{}
		holder.Store(db)
		server.dbSet[i] = holder
	}
	return server
}
