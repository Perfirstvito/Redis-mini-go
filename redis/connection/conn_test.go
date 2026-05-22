package connection

import (
	"testing"
)

// TestSelectDB 测试数据库切换
func TestSelectDB(t *testing.T) {
	c := NewFakeConn()
	c.SelectDB(3)
	if c.GetDBIndex() != 3 {
		t.Errorf("expected 3, got %d", c.GetDBIndex())
	}
}

// TestSelectDBDefault 测试默认数据库是 0
func TestSelectDBDefault(t *testing.T) {
	c := NewFakeConn()
	if c.GetDBIndex() != 0 {
		t.Errorf("expected default db 0, got %d", c.GetDBIndex())
	}
}

// TestPassword 测试密码存取
func TestPassword(t *testing.T) {
	c := NewFakeConn()
	c.SetPassword("mypassword")
	if c.GetPassword() != "mypassword" {
		t.Errorf("expected mypassword, got %s", c.GetPassword())
	}
}

// TestMultiState 测试事务状态切换
func TestMultiState(t *testing.T) {
	c := NewFakeConn()

	// 默认不在事务中
	if c.InMultiState() {
		t.Errorf("expected not in multi state initially")
	}

	// 进入事务
	c.SetMultiState(true)
	if !c.InMultiState() {
		t.Errorf("expected in multi state after SetMultiState(true)")
	}

	// 退出事务
	c.SetMultiState(false)
	if c.InMultiState() {
		t.Errorf("expected not in multi state after SetMultiState(false)")
	}
}

// TestEnqueueCmd 测试命令入队
func TestEnqueueCmd(t *testing.T) {
	c := NewFakeConn()
	c.SetMultiState(true)

	cmd1 := [][]byte{[]byte("SET"), []byte("a"), []byte("1")}
	cmd2 := [][]byte{[]byte("INCR"), []byte("a")}

	c.EnqueueCmd(cmd1)
	c.EnqueueCmd(cmd2)

	queue := c.GetQueuedCmdLine()
	if len(queue) != 2 {
		t.Errorf("expected 2 cmds in queue, got %d", len(queue))
	}

	// 验证第一个命令
	if string(queue[0][0]) != "SET" {
		t.Errorf("expected first cmd SET, got %s", string(queue[0][0]))
	}
}

// TestClearQueuedCmds 测试清空队列
func TestClearQueuedCmds(t *testing.T) {
	c := NewFakeConn()
	c.SetMultiState(true)

	c.EnqueueCmd([][]byte{[]byte("SET"), []byte("a"), []byte("1")})
	c.EnqueueCmd([][]byte{[]byte("SET"), []byte("b"), []byte("2")})

	c.ClearQueuedCmds()

	queue := c.GetQueuedCmdLine()
	if len(queue) != 0 {
		t.Errorf("expected empty queue after ClearQueuedCmds, got %d", len(queue))
	}
}

// TestSubscribe 测试订阅功能
func TestSubscribe(t *testing.T) {
	c := NewFakeConn()

	c.Subscribe("news")
	c.Subscribe("sports")

	if c.SubsCount() != 2 {
		t.Errorf("expected 2 subscriptions, got %d", c.SubsCount())
	}

	channels := c.GetChannels()
	if len(channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(channels))
	}
}

// TestUnsubscribe 测试取消订阅
func TestUnsubscribe(t *testing.T) {
	c := NewFakeConn()

	c.Subscribe("news")
	c.Subscribe("sports")
	c.UnSubscribe("news")

	if c.SubsCount() != 1 {
		t.Errorf("expected 1 subscription after unsubscribe, got %d", c.SubsCount())
	}
}

// TestWatching 测试 WATCH 乐观锁
func TestWatching(t *testing.T) {
	c := NewFakeConn()

	watching := c.GetWatching()
	watching["mykey"] = 1
	watching["otherkey"] = 2

	// 再次获取应该返回同一个 map
	watching2 := c.GetWatching()
	if len(watching2) != 2 {
		t.Errorf("expected 2 watching keys, got %d", len(watching2))
	}
}

// TestTxErrors 测试事务错误记录
func TestTxErrors(t *testing.T) {
	c := NewFakeConn()

	err1 := &errRetryTx{} // 模拟错误
	err2 := &errRetryTx{}

	c.AddTxError(err1)
	c.AddTxError(err2)

	errors := c.GetTxErrors()
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}
}

// TestFakeConnWriteRead 测试 FakeConn 的 Write/Read
func TestFakeConnWriteRead(t *testing.T) {
	c := NewFakeConn()

	// 写入数据
	data := []byte("hello world")
	n, err := c.Write(data)
	if err != nil {
		t.Errorf("write error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected write %d bytes, got %d", len(data), n)
	}

	// 读取数据
	buf := make([]byte, 1024)
	n, err = c.Read(buf)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected read %d bytes, got %d", len(data), n)
	}
	if string(buf[:n]) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(buf[:n]))
	}
}

// TestFakeConnClose 测试 FakeConn 关闭
func TestFakeConnClose(t *testing.T) {
	c := NewFakeConn()
	err := c.Close()
	if err != nil {
		t.Errorf("close error: %v", err)
	}

	// 关闭后写入应该返回 EOF
	_, err = c.Write([]byte("test"))
	if err == nil {
		t.Errorf("expected EOF after close, got nil")
	}
}

// TestName 测试连接名称
func TestName(t *testing.T) {
	c := NewFakeConn()
	name := c.Name()
	// FakeConn 返回空字符串
	if name != "" {
		t.Errorf("expected empty name for FakeConn, got %s", name)
	}
}

// errRetryTx 模拟事务错误
type errRetryTx struct{}

func (e *errRetryTx) Error() string {
	return "retry transaction"
}
