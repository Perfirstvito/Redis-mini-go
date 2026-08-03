package database

import (
	"sync/atomic"
	"testing"

	"my-redis/config"
	"my-redis/interfaces/redis"
	"my-redis/pubsub"
	"my-redis/redis/connection"
)

func TestServerExecRoutesCommandsAcrossDatabases(t *testing.T) {
	setServerTestPassword(t, "")
	server := makeUnitServer(2)
	conn := connection.NewFakeConn()

	assertServerReply(t, server.Exec(conn, serverCmdLine("ping")), "+PONG\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "key", "db0")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("select", "1")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("get", "key")), "$-1\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "key", "db1")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("dbsize")), ":1\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("select", "0")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("get", "key")), "$3\r\ndb0\r\n")
}

func TestServerExecRequiresAuthentication(t *testing.T) {
	setServerTestPassword(t, "secret")
	server := makeUnitServer(1)
	conn := connection.NewFakeConn()

	assertServerReply(t, server.Exec(conn, serverCmdLine("ping")), "+PONG\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "key", "value")), "-NOAUTH Authentication required\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("auth", "wrong")), "-ERR invalid password\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("auth", "secret")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "key", "value")), "+OK\r\n")
}

func TestServerExecRejectsWritesOnReadOnlySlave(t *testing.T) {
	setServerTestPassword(t, "")
	server := makeUnitServer(1)
	client := connection.NewFakeConn()

	assertServerReply(t, server.Exec(client, serverCmdLine("set", "key", "old")), "+OK\r\n")
	atomic.StoreInt32(&server.role, slaveRole)

	assertServerReply(t, server.Exec(client, serverCmdLine("get", "key")), "$3\r\nold\r\n")
	assertServerReply(t, server.Exec(client, serverCmdLine("set", "key", "new")), "-READONLY You can't write against a read only slave.\r\n")

	master := connection.NewFakeConn()
	master.SetMaster()
	assertServerReply(t, server.Exec(master, serverCmdLine("set", "key", "new")), "+OK\r\n")
	assertServerReply(t, server.Exec(client, serverCmdLine("get", "key")), "$3\r\nnew\r\n")
}

func TestServerFlushDBReplacesOnlySelectedDatabase(t *testing.T) {
	setServerTestPassword(t, "")
	server := makeUnitServer(2)
	conn := connection.NewFakeConn()

	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "db0-key", "value")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("select", "1")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("set", "db1-key", "value")), "+OK\r\n")

	oldDB := server.mustSelectDB(1)
	aofCalls := 0
	oldDB.addAof = func(CmdLine) {
		aofCalls++
	}

	assertServerReply(t, server.Exec(conn, serverCmdLine("flushdb")), "+OK\r\n")
	newDB := server.mustSelectDB(1)
	if newDB == oldDB {
		t.Fatal("flushdb should replace the selected database")
	}
	newDB.addAof(serverCmdLine("set", "probe", "value"))
	if aofCalls != 1 {
		t.Fatalf("replacement database AOF callback calls = %d, want 1", aofCalls)
	}
	assertServerReply(t, server.Exec(conn, serverCmdLine("get", "db1-key")), "$-1\r\n")

	assertServerReply(t, server.Exec(conn, serverCmdLine("select", "0")), "+OK\r\n")
	assertServerReply(t, server.Exec(conn, serverCmdLine("get", "db0-key")), "$5\r\nvalue\r\n")
	assertServerReply(t, server.flushDB(2), "-ERR DB index is out of range\r\n")
}

func TestServerAfterClientCloseRemovesSubscriptions(t *testing.T) {
	setServerTestPassword(t, "")
	server := makeUnitServer(1)
	conn := connection.NewFakeConn()

	assertServerReply(t, server.Exec(conn, serverCmdLine("subscribe", "news")), "")
	if got := conn.SubsCount(); got != 1 {
		t.Fatalf("subscription count = %d, want 1", got)
	}

	server.AfterClientClose(conn)
	if got := conn.SubsCount(); got != 0 {
		t.Fatalf("subscription count after close = %d, want 0", got)
	}
	assertServerReply(t, server.Exec(conn, serverCmdLine("publish", "news", "message")), ":0\r\n")
}

func makeUnitServer(dbCount int) *Server {
	server := &Server{
		dbSet:       make([]*atomic.Value, dbCount),
		hub:         pubsub.MakeHub(),
		role:        masterRole,
		slaveStatus: initReplSlaveStatus(),
		slogLogger:  NewSlowLogger(16, 0),
	}
	for i := range server.dbSet {
		db := makeDB()
		db.index = i
		holder := &atomic.Value{}
		holder.Store(db)
		server.dbSet[i] = holder
	}
	server.initMasterStatus()
	return server
}

func setServerTestPassword(t *testing.T, password string) {
	t.Helper()
	oldProperties := config.Properties
	properties := *oldProperties
	properties.RequirePass = password
	config.Properties = &properties
	t.Cleanup(func() {
		config.Properties = oldProperties
	})
}

func serverCmdLine(args ...string) CmdLine {
	line := make(CmdLine, len(args))
	for i, arg := range args {
		line[i] = []byte(arg)
	}
	return line
}

func assertServerReply(t *testing.T, reply redis.Reply, want string) {
	t.Helper()
	if got := string(reply.ToBytes()); got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}
