package pubsub

import (
	"bytes"
	"testing"

	"my-redis/datastruct/list"
	"my-redis/redis/connection"
	"my-redis/redis/protocol"
)

// cmdArgs builds a [][]byte command tail the way the redis server passes it.
func cmdArgs(parts ...string) [][]byte {
	args := make([][]byte, len(parts))
	for i, p := range parts {
		args[i] = []byte(p)
	}
	return args
}

func TestMakeHubInitializesHub(t *testing.T) {
	hub := MakeHub()
	if hub == nil {
		t.Fatal("MakeHub() returned nil")
	}
	if hub.subs == nil {
		t.Fatal("hub.subs dict not initialized")
	}
	if hub.subsLocker == nil {
		t.Fatal("hub.subsLocker not initialized")
	}
}

func TestMakeMsgFormat(t *testing.T) {
	got := makeMsg("subscribe", "news", 1)
	want := []byte("*3\r\n$9\r\nsubscribe\r\n$4\r\nnews\r\n:1\r\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("makeMsg() = %q, want %q", got, want)
	}
}

func TestSubscribeAddsClientAndNotifies(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	reply := Subscribe(hub, c, cmdArgs("news"))
	if _, ok := reply.(*protocol.NoReply); !ok {
		t.Fatalf("Subscribe reply = %T, want *NoReply", reply)
	}
	if c.SubsCount() != 1 {
		t.Fatalf("SubsCount() = %d, want 1", c.SubsCount())
	}

	// the subscribe reply was written to the connection
	wantMsg := makeMsg(_subscribe, "news", 1)
	if !bytes.Equal(c.Bytes(), wantMsg) {
		t.Fatalf("conn buffer = %q, want %q", c.Bytes(), wantMsg)
	}

	// hub now records the subscriber
	raw, ok := hub.subs.Get("news")
	if !ok {
		t.Fatal("hub.subs missing channel after Subscribe")
	}
	subscribers, _ := raw.(*list.LinkedList)
	if subscribers.Len() != 1 {
		t.Fatalf("subscribers len = %d, want 1", subscribers.Len())
	}
}

func TestSubscribeMultipleChannels(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	Subscribe(hub, c, cmdArgs("news", "sports"))

	if c.SubsCount() != 2 {
		t.Fatalf("SubsCount() = %d, want 2", c.SubsCount())
	}
	channels := c.GetChannels()
	if len(channels) != 2 {
		t.Fatalf("GetChannels() len = %d, want 2", len(channels))
	}
}

func TestSubscribeDuplicateChannelWritesOnce(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	Subscribe(hub, c, cmdArgs("news"))
	firstLen := len(c.Bytes())
	Subscribe(hub, c, cmdArgs("news"))

	if c.SubsCount() != 1 {
		t.Fatalf("SubsCount() = %d, want 1 after duplicate subscribe", c.SubsCount())
	}
	if len(c.Bytes()) != firstLen {
		t.Fatalf("buffer grew from %d to %d; duplicate subscribe should not write", firstLen, len(c.Bytes()))
	}

	// subscriber list still has exactly one entry
	raw, _ := hub.subs.Get("news")
	subscribers, _ := raw.(*list.LinkedList)
	if subscribers.Len() != 1 {
		t.Fatalf("subscribers len = %d, want 1", subscribers.Len())
	}
}

func TestUnSubscribeRemovesClient(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	Subscribe(hub, c, cmdArgs("news"))
	c.Clean() // drop subscribe reply so we can inspect the unsubscribe reply alone

	UnSubscribe(hub, c, cmdArgs("news"))

	if c.SubsCount() != 0 {
		t.Fatalf("SubsCount() = %d, want 0 after unsubscribe", c.SubsCount())
	}
	wantMsg := makeMsg(_unsubscribe, "news", 0)
	if !bytes.Equal(c.Bytes(), wantMsg) {
		t.Fatalf("conn buffer = %q, want %q", c.Bytes(), wantMsg)
	}
	if _, ok := hub.subs.Get("news"); ok {
		t.Fatal("hub.subs should drop channel once it has no subscribers")
	}
}

func TestUnSubscribeNoArgsUnsubscribesAllChannels(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	Subscribe(hub, c, cmdArgs("news", "sports"))
	c.Clean()

	UnSubscribe(hub, c, nil)

	if c.SubsCount() != 0 {
		t.Fatalf("SubsCount() = %d, want 0", c.SubsCount())
	}
	if _, ok := hub.subs.Get("news"); ok {
		t.Fatal("hub.subs still has news")
	}
	if _, ok := hub.subs.Get("sports"); ok {
		t.Fatal("hub.subs still has sports")
	}
}

func TestUnSubscribeWithNoChannelsWritesNothing(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	UnSubscribe(hub, c, nil)

	if !bytes.Equal(c.Bytes(), unSubscribeNothing) {
		t.Fatalf("conn buffer = %q, want unSubscribeNothing", c.Bytes())
	}
}

func TestUnsubscribeAllRemovesFromEveryChannel(t *testing.T) {
	hub := MakeHub()
	c := connection.NewFakeConn()

	Subscribe(hub, c, cmdArgs("news", "sports"))

	UnsubscribeAll(hub, c)

	if c.SubsCount() != 0 {
		t.Fatalf("SubsCount() = %d, want 0", c.SubsCount())
	}
	if _, ok := hub.subs.Get("news"); ok {
		t.Fatal("hub.subs still has news after UnsubscribeAll")
	}
	if _, ok := hub.subs.Get("sports"); ok {
		t.Fatal("hub.subs still has sports after UnsubscribeAll")
	}
}

func TestPublishDeliversToSubscribers(t *testing.T) {
	hub := MakeHub()
	sub1 := connection.NewFakeConn()
	sub2 := connection.NewFakeConn()

	Subscribe(hub, sub1, cmdArgs("news"))
	Subscribe(hub, sub2, cmdArgs("news"))
	sub1.Clean()
	sub2.Clean()

	reply := Publish(hub, cmdArgs("news", "hello"))
	if got, want := string(reply.ToBytes()), ":2\r\n"; got != want {
		t.Fatalf("Publish reply = %q, want %q", got, want)
	}

	wantPayload := []byte("*3\r\n$7\r\nmessage\r\n$4\r\nnews\r\n$5\r\nhello\r\n")
	if !bytes.Equal(sub1.Bytes(), wantPayload) {
		t.Fatalf("sub1 buffer = %q, want %q", sub1.Bytes(), wantPayload)
	}
	if !bytes.Equal(sub2.Bytes(), wantPayload) {
		t.Fatalf("sub2 buffer = %q, want %q", sub2.Bytes(), wantPayload)
	}
}

func TestPublishNoSubscribersReturnsZero(t *testing.T) {
	hub := MakeHub()
	reply := Publish(hub, cmdArgs("ghost", "hello"))
	if got, want := string(reply.ToBytes()), ":0\r\n"; got != want {
		t.Fatalf("Publish reply = %q, want %q", got, want)
	}
}

func TestPublishWrongArgCount(t *testing.T) {
	hub := MakeHub()
	reply := Publish(hub, cmdArgs("news"))
	if got, want := string(reply.ToBytes()), "-ERR wrong number of arguments for 'publish' command\r\n"; got != want {
		t.Fatalf("Publish reply = %q, want %q", got, want)
	}
}

func TestPublishOnlyDeliversToMatchingChannel(t *testing.T) {
	hub := MakeHub()
	news := connection.NewFakeConn()
	sports := connection.NewFakeConn()

	Subscribe(hub, news, cmdArgs("news"))
	Subscribe(hub, sports, cmdArgs("sports"))
	news.Clean()
	sports.Clean()

	Publish(hub, cmdArgs("news", "hello"))

	if !bytes.Contains(news.Bytes(), []byte("hello")) {
		t.Fatalf("news subscriber should receive the message, got %q", news.Bytes())
	}
	if len(sports.Bytes()) != 0 {
		t.Fatalf("sports subscriber should not receive message, got %q", sports.Bytes())
	}
}
