package protocol

import "testing"

func TestFixedRepliesToBytes(t *testing.T) {
	tests := []struct {
		name  string
		reply interface {
			ToBytes() []byte
		}
		want string
	}{
		{name: "pong", reply: &PongReply{}, want: "+PONG\r\n"},
		{name: "ok", reply: MakeOkReply(), want: "+OK\r\n"},
		{name: "null bulk", reply: MakeNullBulkReply(), want: "$-1\r\n"},
		{name: "empty multi bulk", reply: MakeEmptyMultiBulkReply(), want: "*0\r\n"},
		{name: "no reply", reply: &NoReply{}, want: ""},
		{name: "queued", reply: MakeQueuedReply(), want: "+QUEUED\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.reply.ToBytes()); got != tt.want {
				t.Fatalf("ToBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReplyFactoriesReuseSharedInstances(t *testing.T) {
	if MakeOkReply() != MakeOkReply() {
		t.Fatal("MakeOkReply should return the shared OK reply instance")
	}
	if MakeQueuedReply() != MakeQueuedReply() {
		t.Fatal("MakeQueuedReply should return the shared QUEUED reply instance")
	}
}

func TestIsEmptyMultiBulkReply(t *testing.T) {
	if !IsEmptyMultiBulkReply(MakeEmptyMultiBulkReply()) {
		t.Fatal("empty multi bulk reply should be detected")
	}
	if IsEmptyMultiBulkReply(MakeBulkReply([]byte{})) {
		t.Fatal("empty bulk string should not be detected as empty multi bulk reply")
	}
	if IsEmptyMultiBulkReply(&NoReply{}) {
		t.Fatal("no reply should not be detected as empty multi bulk reply")
	}
}

func TestNilBulkReplyUsesNullBulkReplyBytes(t *testing.T) {
	if got, want := string(MakeBulkReply(nil).ToBytes()), "$-1\r\n"; got != want {
		t.Fatalf("nil bulk reply = %q, want %q", got, want)
	}
}
