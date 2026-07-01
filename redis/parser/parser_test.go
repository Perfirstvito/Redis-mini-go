package parser

import (
	"errors"
	"io"
	"my-redis/interfaces/redis"
	"strings"
	"testing"
)

func TestParseBytesParsesRESPReplies(t *testing.T) {
	input := strings.Join([]string{
		"+OK\r\n",
		"-ERR wrong\r\n",
		":42\r\n",
		"$5\r\nhello\r\n",
		"*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
		"PING world\r\n",
	}, "")

	replies, err := ParseBytes([]byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() returned error: %v", err)
	}

	assertReplyBytes(t, replies, []string{
		"+OK\r\n",
		"-ERR wrong\r\n",
		":42\r\n",
		"$5\r\nhello\r\n",
		"*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
		"*2\r\n$4\r\nPING\r\n$5\r\nworld\r\n",
	})
}

func TestParseBytesParsesNullAndEmptyReplies(t *testing.T) {
	replies, err := ParseBytes([]byte("$-1\r\n$0\r\n\r\n*0\r\n"))
	if err != nil {
		t.Fatalf("ParseBytes() returned error: %v", err)
	}

	assertReplyBytes(t, replies, []string{
		"$-1\r\n",
		"$0\r\n\r\n",
		"*0\r\n",
	})
}

func TestParseOneReturnsFirstReply(t *testing.T) {
	reply, err := ParseOne([]byte("+OK\r\n+PONG\r\n"))
	if err != nil {
		t.Fatalf("ParseOne() returned error: %v", err)
	}
	if got, want := string(reply.ToBytes()), "+OK\r\n"; got != want {
		t.Fatalf("ParseOne() = %q, want %q", got, want)
	}
}

func TestParseStreamReportsEOF(t *testing.T) {
	ch := ParseStream(strings.NewReader("+OK\r\n"))

	payload := <-ch
	if payload == nil || payload.Err != nil {
		t.Fatalf("first payload = %#v, want OK reply", payload)
	}
	if got, want := string(payload.Data.ToBytes()), "+OK\r\n"; got != want {
		t.Fatalf("first reply = %q, want %q", got, want)
	}

	payload = <-ch
	if payload == nil || !errors.Is(payload.Err, io.EOF) {
		t.Fatalf("second payload = %#v, want EOF error", payload)
	}

	if payload, ok := <-ch; ok || payload != nil {
		t.Fatalf("channel should be closed after EOF, got payload=%#v ok=%v", payload, ok)
	}
}

func TestParseBytesIgnoresEmptyLines(t *testing.T) {
	replies, err := ParseBytes([]byte("\r\n+OK\r\n"))
	if err != nil {
		t.Fatalf("ParseBytes() returned error: %v", err)
	}

	assertReplyBytes(t, replies, []string{"+OK\r\n"})
}

func TestParseBytesReturnsProtocolErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "illegal integer",
			input:   ":not-number\r\n",
			wantErr: "protocol error: illegal number not-number",
		},
		{
			name:    "illegal bulk header",
			input:   "$-2\r\n",
			wantErr: "protocol error: illegal bulk string header: $-2",
		},
		{
			name:    "illegal array header",
			input:   "*x\r\n",
			wantErr: "protocol error: illegal array header x",
		},
		{
			name:    "illegal array element header",
			input:   "*1\r\n+OK\r\n",
			wantErr: "protocol error: illegal bulk string header +OK\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replies, err := ParseBytes([]byte(tt.input))
			if err == nil {
				t.Fatalf("ParseBytes() error = nil, replies = %#v", replies)
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("ParseBytes() error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestParseBytesReturnsUnexpectedEOFForTruncatedBulkString(t *testing.T) {
	replies, err := ParseBytes([]byte("$5\r\nhel"))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("ParseBytes() error = %v, want unexpected EOF; replies = %#v", err, replies)
	}
}

func TestParseBytesParsesFullResyncRDBBulkWithoutTrailingCRLF(t *testing.T) {
	replies, err := ParseBytes([]byte("+FULLRESYNC runid 0\r\n$4\r\nRDB!+OK\r\n"))
	if err != nil {
		t.Fatalf("ParseBytes() returned error: %v", err)
	}

	assertReplyBytes(t, replies, []string{
		"+FULLRESYNC runid 0\r\n",
		"$4\r\nRDB!\r\n",
		"+OK\r\n",
	})
}

func assertReplyBytes(t *testing.T, replies []redis.Reply, want []string) {
	t.Helper()
	if len(replies) != len(want) {
		t.Fatalf("reply count = %d, want %d", len(replies), len(want))
	}
	for i, reply := range replies {
		if got := string(reply.ToBytes()); got != want[i] {
			t.Fatalf("reply[%d] = %q, want %q", i, got, want[i])
		}
	}
}
