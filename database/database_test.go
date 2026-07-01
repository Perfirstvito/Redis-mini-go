package database

import (
	"testing"
	"time"

	idatabase "my-redis/interfaces/database"
	"my-redis/interfaces/redis"
	"my-redis/redis/protocol"
)

func TestMakeDBInitializesStorage(t *testing.T) {
	db := makeDB()
	if db == nil {
		t.Fatal("makeDB() returned nil")
	}
	if db.data == nil || db.ttlMap == nil || db.versionMap == nil {
		t.Fatalf("makeDB() did not initialize storage maps: %#v", db)
	}
	if db.addAof == nil {
		t.Fatal("makeDB() did not initialize addAof")
	}

	db.addAof(CmdLine{[]byte("PING")})
}

func TestPutGetRemoveEntityAndCallbacks(t *testing.T) {
	db := makeDB()
	db.index = 2

	var inserted []string
	db.insertCallback = func(dbIndex int, key string, entity *idatabase.DataEntity) {
		inserted = append(inserted, key)
		if dbIndex != 2 {
			t.Fatalf("insert callback dbIndex = %d, want 2", dbIndex)
		}
	}

	var deletedKey string
	var deletedEntity *idatabase.DataEntity
	db.deleteCallback = func(dbIndex int, key string, entity *idatabase.DataEntity) {
		if dbIndex != 2 {
			t.Fatalf("delete callback dbIndex = %d, want 2", dbIndex)
		}
		deletedKey = key
		deletedEntity = entity
	}

	first := &idatabase.DataEntity{Data: "first"}
	if got := db.PutEntity("key", first); got != 1 {
		t.Fatalf("PutEntity(new) = %d, want 1", got)
	}
	if len(inserted) != 1 || inserted[0] != "key" {
		t.Fatalf("insert callback keys = %v, want [key]", inserted)
	}

	replacement := &idatabase.DataEntity{Data: "replacement"}
	if got := db.PutEntity("key", replacement); got != 0 {
		t.Fatalf("PutEntity(existing) = %d, want 0", got)
	}
	if len(inserted) != 1 {
		t.Fatalf("insert callback should not run for replacement, got %d calls", len(inserted))
	}

	gotEntity, ok := db.GetEntity("key")
	if !ok || gotEntity != replacement {
		t.Fatalf("GetEntity() = (%#v, %v), want replacement and ok=true", gotEntity, ok)
	}

	edited := &idatabase.DataEntity{Data: "edited"}
	if got := db.PutIfExists("key", edited); got != 1 {
		t.Fatalf("PutIfExists(existing) = %d, want 1", got)
	}
	if got := db.PutIfExists("missing", first); got != 0 {
		t.Fatalf("PutIfExists(missing) = %d, want 0", got)
	}

	absent := &idatabase.DataEntity{Data: "absent"}
	if got := db.PutIfAbsent("key", absent); got != 0 {
		t.Fatalf("PutIfAbsent(existing) = %d, want 0", got)
	}
	if got := db.PutIfAbsent("new-key", absent); got != 1 {
		t.Fatalf("PutIfAbsent(missing) = %d, want 1", got)
	}

	db.Remove("key")
	if deletedKey != "key" || deletedEntity != edited {
		t.Fatalf("delete callback = (%q, %#v), want key and edited entity", deletedKey, deletedEntity)
	}
	if gotEntity, ok := db.GetEntity("key"); ok || gotEntity != nil {
		t.Fatalf("GetEntity(removed) = (%#v, %v), want nil and ok=false", gotEntity, ok)
	}
}

func TestRemovesReturnsDeletedCount(t *testing.T) {
	db := makeDB()
	db.PutEntity("a", &idatabase.DataEntity{Data: "a"})
	db.PutEntity("b", &idatabase.DataEntity{Data: "b"})

	if got := db.Removes("a", "missing", "b"); got != 2 {
		t.Fatalf("Removes() = %d, want 2", got)
	}
	if got := db.data.Len(); got != 0 {
		t.Fatalf("data len = %d, want 0", got)
	}
}

func TestExpirationLifecycle(t *testing.T) {
	db := makeDB()
	db.PutEntity("key", &idatabase.DataEntity{Data: "value"})

	expireAt := time.Now().Add(time.Hour)
	db.Expire("key", expireAt)
	t.Cleanup(func() {
		db.Persist("key")
	})

	rawExpireAt, ok := db.ttlMap.Get("key")
	if !ok {
		t.Fatal("Expire() did not write ttlMap")
	}
	if got := rawExpireAt.(time.Time); !got.Equal(expireAt) {
		t.Fatalf("Expire() stored %v, want %v", got, expireAt)
	}
	if db.IsExpired("key") {
		t.Fatal("key with future expiration should not be expired")
	}

	db.ttlMap.Put("key", time.Now().Add(-time.Second))
	if !db.IsExpired("key") {
		t.Fatal("key with past expiration should be expired")
	}
	if gotEntity, ok := db.GetEntity("key"); ok || gotEntity != nil {
		t.Fatalf("expired key entity = (%#v, %v), want nil and ok=false", gotEntity, ok)
	}

	db.PutEntity("persist-key", &idatabase.DataEntity{Data: "persist"})
	db.ttlMap.Put("persist-key", time.Now().Add(time.Hour))
	db.Persist("persist-key")
	if _, ok := db.ttlMap.Get("persist-key"); ok {
		t.Fatal("Persist() should remove ttlMap entry")
	}
}

func TestVersionLifecycle(t *testing.T) {
	db := makeDB()
	if got := db.GetVersion("key"); got != 0 {
		t.Fatalf("initial version = %d, want 0", got)
	}

	db.addVersion("key", "key", "other")

	if got := db.GetVersion("key"); got != 2 {
		t.Fatalf("version for key = %d, want 2", got)
	}
	if got := db.GetVersion("other"); got != 1 {
		t.Fatalf("version for other = %d, want 1", got)
	}
}

func TestForEachIncludesEntityAndExpiration(t *testing.T) {
	db := makeDB()
	entity := &idatabase.DataEntity{Data: "value"}
	expireAt := time.Now().Add(time.Hour)

	db.PutEntity("key", entity)
	db.ttlMap.Put("key", expireAt)

	seen := false
	db.ForEach(func(key string, data *idatabase.DataEntity, expiration *time.Time) bool {
		seen = key == "key" && data == entity && expiration != nil && expiration.Equal(expireAt)
		return true
	})
	if !seen {
		t.Fatal("ForEach did not expose the expected key, entity, and expiration")
	}
}

func TestValidateArity(t *testing.T) {
	tests := []struct {
		name    string
		arity   int
		cmdLine CmdLine
		want    bool
	}{
		{name: "exact match", arity: 2, cmdLine: CmdLine{[]byte("get"), []byte("key")}, want: true},
		{name: "exact mismatch", arity: 2, cmdLine: CmdLine{[]byte("get")}, want: false},
		{name: "minimum match", arity: -2, cmdLine: CmdLine{[]byte("del"), []byte("key")}, want: true},
		{name: "minimum allows more", arity: -2, cmdLine: CmdLine{[]byte("del"), []byte("a"), []byte("b")}, want: true},
		{name: "minimum mismatch", arity: -2, cmdLine: CmdLine{[]byte("del")}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateArity(tt.arity, tt.cmdLine); got != tt.want {
				t.Fatalf("validateArity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecNormalCommandDispatchesRegisteredCommand(t *testing.T) {
	db := makeDB()
	restoreCommandTable(t)

	registerCommand("TESTWRITE", func(db *DB, args [][]byte) redis.Reply {
		return protocol.MakeStatusReply("wrote " + string(args[0]))
	}, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, nil, 2)

	reply := db.execNormalCommand(CmdLine{[]byte("TESTWRITE"), []byte("key")})
	if got, want := string(reply.ToBytes()), "+wrote key\r\n"; got != want {
		t.Fatalf("execNormalCommand() = %q, want %q", got, want)
	}
	if got := db.GetVersion("key"); got != 1 {
		t.Fatalf("version for written key = %d, want 1", got)
	}
}

func TestExecNormalCommandReturnsErrorReply(t *testing.T) {
	db := makeDB()
	restoreCommandTable(t)

	if got, want := string(db.execNormalCommand(CmdLine{[]byte("missing")}).ToBytes()), "-ERR unknown command 'missing'\r\n"; got != want {
		t.Fatalf("unknown command reply = %q, want %q", got, want)
	}

	registerCommand("onearg", func(db *DB, args [][]byte) redis.Reply {
		return protocol.MakeOkReply()
	}, nil, nil, 2)

	if got, want := string(db.execNormalCommand(CmdLine{[]byte("onearg")}).ToBytes()), "-ERR wrong number of arguments for 'onearg' command\r\n"; got != want {
		t.Fatalf("wrong arity reply = %q, want %q", got, want)
	}
}

func TestExecWithLockDispatchesWithoutUpdatingVersion(t *testing.T) {
	db := makeDB()
	restoreCommandTable(t)

	registerCommand("read", func(db *DB, args [][]byte) redis.Reply {
		return protocol.MakeStatusReply("read " + string(args[0]))
	}, nil, nil, 2)

	reply := db.execWithLock(CmdLine{[]byte("read"), []byte("key")})
	if got, want := string(reply.ToBytes()), "+read key\r\n"; got != want {
		t.Fatalf("execWithLock() = %q, want %q", got, want)
	}
	if got := db.GetVersion("key"); got != 0 {
		t.Fatalf("execWithLock should not update key version, got %d", got)
	}
}

func restoreCommandTable(t *testing.T) {
	t.Helper()
	old := cmdTable
	cmdTable = make(map[string]*command)
	t.Cleanup(func() {
		cmdTable = old
	})
}
