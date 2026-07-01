package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// saveProperties snapshots the global Properties and restores it after the test,
// so config-mutating tests stay isolated.
func saveProperties(t *testing.T) {
	t.Helper()
	savedProps := Properties
	savedPath := configFilePath
	t.Cleanup(func() {
		Properties = savedProps
		configFilePath = savedPath
	})
}

func TestParseReadsTypedFields(t *testing.T) {
	saveProperties(t)
	src := strings.NewReader(strings.Join([]string{
		"bind 0.0.0.0",
		"port 7000",
		"appendonly yes",
		"slowlog-log-slower-than 20000",
		"slowlog-max-len 128",
		"databases 16",
		"cluster-enable yes",
	}, "\n"))
	cfg := parse(src)

	if cfg.Bind != "0.0.0.0" {
		t.Fatalf("Bind = %q, want 0.0.0.0", cfg.Bind)
	}
	if cfg.Port != 7000 {
		t.Fatalf("Port = %d, want 7000", cfg.Port)
	}
	if !cfg.AppendOnly {
		t.Fatalf("AppendOnly = false, want true")
	}
	if cfg.SlowLogSlowerThan != 20000 {
		t.Fatalf("SlowLogSlowerThan = %d, want 20000", cfg.SlowLogSlowerThan)
	}
	if cfg.SlowLogMaxLen != 128 {
		t.Fatalf("SlowLogMaxLen = %d, want 128", cfg.SlowLogMaxLen)
	}
	if cfg.Databases != 16 {
		t.Fatalf("Databases = %d, want 16", cfg.Databases)
	}
	if !cfg.ClusterEnable {
		t.Fatalf("ClusterEnable = false, want true")
	}
}

func TestParseBoolOnlyTrueForYes(t *testing.T) {
	saveProperties(t)
	tests := []struct {
		value string
		want  bool
	}{
		{"yes", true},
		{"no", false},
		// redis bools use yes/no, not Go's ParseBool semantics
		{"true", false},
		{"1", false},
		{"", false},
	}
	for _, tt := range tests {
		cfg := parse(strings.NewReader("appendonly " + tt.value))
		if cfg.AppendOnly != tt.want {
			t.Fatalf("appendonly %q -> AppendOnly = %v, want %v", tt.value, cfg.AppendOnly, tt.want)
		}
	}
}

func TestParseIgnoresCommentsAndLinesWithoutSeparator(t *testing.T) {
	saveProperties(t)
	src := strings.NewReader(strings.Join([]string{
		"# this is a comment",
		"",
		"   # indented comment",
		"port 6380",
		"lonelytoken", // no space separator -> skipped, must not panic
	}, "\n"))
	cfg := parse(src)

	if cfg.Port != 6380 {
		t.Fatalf("Port = %d, want 6380", cfg.Port)
	}
	if cfg.Bind != "" {
		t.Fatalf("Bind = %q, want empty (unset)", cfg.Bind)
	}
}

func TestParseMatchesKeyCaseInsensitively(t *testing.T) {
	saveProperties(t)
	cfg := parse(strings.NewReader("PORT 9000"))
	if cfg.Port != 9000 {
		t.Fatalf("Port = %d, want 9000 (case-insensitive key)", cfg.Port)
	}
}

func TestParseIntParseErrorLeavesZero(t *testing.T) {
	saveProperties(t)
	cfg := parse(strings.NewReader("port not-a-number"))
	if cfg.Port != 0 {
		t.Fatalf("Port = %d, want 0 on parse error", cfg.Port)
	}
}

func TestParseSkipsUnknownKeys(t *testing.T) {
	saveProperties(t)
	cfg := parse(strings.NewReader("unknown-key value\nport 6390"))
	if cfg.Port != 6390 {
		t.Fatalf("Port = %d, want 6390", cfg.Port)
	}
}

func TestAnnounceAddress(t *testing.T) {
	tests := []struct {
		name string
		p    *ServerProperties
		want string
	}{
		{"uses announce host", &ServerProperties{AnnounceHost: "1.2.3.4", Bind: "127.0.0.1", Port: 6379}, "1.2.3.4:6379"},
		{"falls back to bind", &ServerProperties{AnnounceHost: "", Bind: "127.0.0.1", Port: 6379}, "127.0.0.1:6379"},
		{"zero port", &ServerProperties{Bind: "0.0.0.0", Port: 0}, "0.0.0.0:0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.AnnounceAddress(); got != tt.want {
				t.Fatalf("AnnounceAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRaftAnnounceAddress(t *testing.T) {
	tests := []struct {
		name string
		p    *ServerProperties
		want string
	}{
		{"uses advertise addr", &ServerProperties{RaftAdvertiseAddr: "raft-adv", RaftListenAddr: "raft-listen"}, "raft-adv"},
		{"falls back to listen addr", &ServerProperties{RaftAdvertiseAddr: "", RaftListenAddr: "raft-listen"}, "raft-listen"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.RaftAnnounceAddress(); got != tt.want {
				t.Fatalf("RaftAnnounceAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTmpDir(t *testing.T) {
	saveProperties(t)
	Properties = &ServerProperties{Dir: "/var/godis"}
	if got := GetTmpDir(); got != "/var/godis/tmp" {
		t.Fatalf("GetTmpDir() = %q, want /var/godis/tmp", got)
	}
}

func TestSetupConfigLoadsFileAndSetsPath(t *testing.T) {
	saveProperties(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "redis.conf")
	content := strings.Join([]string{
		"bind 10.0.0.1",
		"port 1234",
		"appendonly yes",
		"databases 8",
	}, "\n")
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	SetupConfig(cfgPath)

	if Properties.Bind != "10.0.0.1" {
		t.Fatalf("Properties.Bind = %q, want 10.0.0.1", Properties.Bind)
	}
	if Properties.Port != 1234 {
		t.Fatalf("Properties.Port = %d, want 1234", Properties.Port)
	}
	if !Properties.AppendOnly {
		t.Fatalf("Properties.AppendOnly = false, want true")
	}
	if Properties.Databases != 8 {
		t.Fatalf("Properties.Databases = %d, want 8", Properties.Databases)
	}
	if Properties.RunID == "" || len(Properties.RunID) != 40 {
		t.Fatalf("Properties.RunID = %q, want 40-char string", Properties.RunID)
	}
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if got := GetConfigFilePath(); got != abs {
		t.Fatalf("GetConfigFilePath() = %q, want %q", got, abs)
	}
	if Properties.Dir != "." {
		t.Fatalf("Properties.Dir = %q, want \".\" (defaulted when unset)", Properties.Dir)
	}
}

func TestSetupConfigPanicsOnMissingFile(t *testing.T) {
	saveProperties(t)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("SetupConfig with missing file should panic")
		}
	}()
	SetupConfig(filepath.Join(t.TempDir(), "no-such.conf"))
}
