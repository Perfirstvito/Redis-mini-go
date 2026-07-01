package protocol

import "testing"

func TestErrorRepliesToBytesAndError(t *testing.T) {
	tests := []struct {
		name  string
		reply interface {
			ToBytes() []byte
			Error() string
		}
		wantBytes string
		wantError string
	}{
		{
			name:      "unknown",
			reply:     &UnknownErrReply{},
			wantBytes: "-Err unknown\r\n",
			wantError: "Err unknown",
		},
		{
			name:      "wrong argument count",
			reply:     MakeArgNumErrReply("get"),
			wantBytes: "-ERR wrong number of arguments for 'get' command\r\n",
			wantError: "ERR wrong number of arguments for 'get' command",
		},
		{
			name:      "syntax",
			reply:     MakeSyntaxErrReply(),
			wantBytes: "-Err syntax error\r\n",
			wantError: "Err syntax error",
		},
		{
			name:      "wrong type",
			reply:     &WrongTypeErrReply{},
			wantBytes: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
			wantError: "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:      "protocol",
			reply:     &ProtocolErrReply{Msg: "bad bulk length"},
			wantBytes: "-ERR Protocol error: 'bad bulk length'\r\n",
			wantError: "ERR Protocol error 'bad bulk length' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.reply.ToBytes()); got != tt.wantBytes {
				t.Fatalf("ToBytes() = %q, want %q", got, tt.wantBytes)
			}
			if got := tt.reply.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestMakeSyntaxErrReplyReusesSharedInstance(t *testing.T) {
	if MakeSyntaxErrReply() != MakeSyntaxErrReply() {
		t.Fatal("MakeSyntaxErrReply should return the shared syntax error reply instance")
	}
}
