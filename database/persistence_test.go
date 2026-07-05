package database

import (
	"strings"
	"testing"

	"my-redis/config"
)

func TestPersistenceLoadRdbFileReportsFilename(t *testing.T) {
	restoreDatabaseConfig(t)
	config.Properties.RDBFilename = t.TempDir() + "/missing.rdb"

	err := (&Server{}).loadRdbFile()
	if err == nil {
		t.Fatal("loadRdbFile() error = nil, want open error")
	}
	if !strings.Contains(err.Error(), "missing.rdb") {
		t.Fatalf("loadRdbFile() error = %q, want filename", err.Error())
	}
}

func TestMakeAuxiliaryServerInitializesDatabases(t *testing.T) {
	restoreDatabaseConfig(t)
	config.Properties.Databases = 2

	server := MakeAuxiliaryServer()
	if len(server.dbSet) != 2 {
		t.Fatalf("dbSet len = %d, want 2", len(server.dbSet))
	}
	for i, holder := range server.dbSet {
		db := holder.Load().(*DB)
		if db == nil || db.data == nil || db.addAof == nil {
			t.Fatalf("db %d not initialized: %#v", i, db)
		}
	}
}

func restoreDatabaseConfig(t *testing.T) {
	t.Helper()
	old := *config.Properties
	t.Cleanup(func() {
		copy := old
		config.Properties = &copy
	})
}
