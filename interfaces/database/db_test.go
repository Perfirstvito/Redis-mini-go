package database

import (
	"testing"
	"time"

	"my-redis/interfaces/redis"

	"github.com/hdt3213/rdb/core"
)

type testReply []byte

func (r testReply) ToBytes() []byte {
	return []byte(r)
}

type testDBEngine struct {
	entity           *DataEntity
	expiration       *time.Time
	insertedCallback KeyEventCallback
	deletedCallback  KeyEventCallback
	closed           bool
}

var _ DB = (*testDBEngine)(nil)
var _ DBEngine = (*testDBEngine)(nil)

func (db *testDBEngine) Exec(client redis.Connection, cmdLine [][]byte) redis.Reply {
	return testReply("+OK\r\n")
}

func (db *testDBEngine) AfterClientClose(c redis.Connection) {}

func (db *testDBEngine) Close() {
	db.closed = true
}

func (db *testDBEngine) LoadRDB(dec *core.Decoder) error {
	return nil
}

func (db *testDBEngine) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {
	return testReply("+LOCKED\r\n")
}

func (db *testDBEngine) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []CmdLine) redis.Reply {
	return testReply("+MULTI\r\n")
}

func (db *testDBEngine) GetUndoLogs(dbIndex int, cmdLine [][]byte) []CmdLine {
	return []CmdLine{cmdLine}
}

func (db *testDBEngine) ForEach(dbIndex int, cb func(key string, data *DataEntity, expiration *time.Time) bool) {
	cb("key", db.entity, db.expiration)
}

func (db *testDBEngine) RWLocks(dbIndex int, writeKeys []string, readKeys []string) {}

func (db *testDBEngine) RWUnLocks(dbIndex int, writeKeys []string, readKeys []string) {}

func (db *testDBEngine) GetDBSize(dbIndex int) (int, int) {
	return 1, 1
}

func (db *testDBEngine) GetEntity(dbIndex int, key string) (*DataEntity, bool) {
	return db.entity, db.entity != nil
}

func (db *testDBEngine) GetExpiration(dbIndex int, key string) *time.Time {
	return db.expiration
}

func (db *testDBEngine) SetKeyInsertedCallback(cb KeyEventCallback) {
	db.insertedCallback = cb
}

func (db *testDBEngine) SetKeyDeletedCallback(cb KeyEventCallback) {
	db.deletedCallback = cb
}

func TestDataEntityStoresArbitraryData(t *testing.T) {
	entity := &DataEntity{
		Data: map[string]int{"count": 3},
	}

	got, ok := entity.Data.(map[string]int)
	if !ok {
		t.Fatalf("Data has type %T, want map[string]int", entity.Data)
	}
	if got["count"] != 3 {
		t.Fatalf("Data[count] = %d, want 3", got["count"])
	}
}

func TestDBEngineContract(t *testing.T) {
	expiration := time.Unix(100, 0)
	entity := &DataEntity{Data: "value"}
	engine := &testDBEngine{
		entity:     entity,
		expiration: &expiration,
	}

	if got := string(engine.Exec(nil, CmdLine{[]byte("PING")}).ToBytes()); got != "+OK\r\n" {
		t.Fatalf("Exec() = %q, want +OK reply", got)
	}

	undoLogs := engine.GetUndoLogs(0, CmdLine{[]byte("SET"), []byte("key"), []byte("value")})
	if len(undoLogs) != 1 || string(undoLogs[0][0]) != "SET" {
		t.Fatalf("GetUndoLogs() = %#v, want original command wrapped in one CmdLine", undoLogs)
	}

	gotEntity, ok := engine.GetEntity(0, "key")
	if !ok || gotEntity != entity {
		t.Fatalf("GetEntity() = (%#v, %v), want stored entity and ok=true", gotEntity, ok)
	}
	if gotExpiration := engine.GetExpiration(0, "key"); gotExpiration != engine.expiration {
		t.Fatalf("GetExpiration() = %v, want %v", gotExpiration, engine.expiration)
	}

	seen := false
	engine.ForEach(0, func(key string, data *DataEntity, gotExpiration *time.Time) bool {
		seen = key == "key" && data == entity && gotExpiration == engine.expiration
		return true
	})
	if !seen {
		t.Fatal("ForEach did not expose stored key, entity, and expiration")
	}

	insertedCalled := false
	engine.SetKeyInsertedCallback(func(dbIndex int, key string, entity *DataEntity) {
		insertedCalled = dbIndex == 0 && key == "key" && entity.Data == "value"
	})
	engine.insertedCallback(0, "key", entity)
	if !insertedCalled {
		t.Fatal("inserted callback was not stored or invoked with expected values")
	}

	engine.Close()
	if !engine.closed {
		t.Fatal("Close should mark the test engine closed")
	}
}
