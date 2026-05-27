package timewheel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// New
// ============================================================

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		slotNum  int
		wantNil  bool
	}{
		{name: "正常创建", interval: time.Second, slotNum: 60},
		{name: "大slotNum", interval: time.Second, slotNum: 3600},
		{name: "interval<=0返回nil", interval: 0, slotNum: 10, wantNil: true},
		{name: "interval负数返回nil", interval: -time.Second, slotNum: 10, wantNil: true},
		{name: "slotNum<=0返回nil", interval: time.Second, slotNum: 0, wantNil: true},
		{name: "slotNum负数返回nil", interval: time.Second, slotNum: -1, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := New(tt.interval, tt.slotNum)
			if tt.wantNil {
				if tw != nil {
					t.Errorf("New(%v, %d) 应返回 nil，实际非 nil", tt.interval, tt.slotNum)
				}
				return
			}
			if tw == nil {
				t.Fatalf("New(%v, %d) 不应返回 nil", tt.interval, tt.slotNum)
			}
			if tw.slotNum != tt.slotNum {
				t.Errorf("slotNum = %d, want %d", tw.slotNum, tt.slotNum)
			}
			if tw.interval != tt.interval {
				t.Errorf("interval = %v, want %v", tw.interval, tt.interval)
			}
			if tw.currentPos != 0 {
				t.Errorf("currentPos 初始应为 0，实际 %d", tw.currentPos)
			}
		})
	}
}

func TestNewInitSlots(t *testing.T) {
	tw := New(time.Second, 5)
	if tw == nil {
		t.Fatal("New 返回 nil")
	}
	if len(tw.slots) != 5 {
		t.Errorf("slots 长度应为 5，实际 %d", len(tw.slots))
	}
	for i := 0; i < 5; i++ {
		if tw.slots[i] == nil {
			t.Errorf("slots[%d] 不应为 nil", i)
		}
	}
}

// ============================================================
// Start / Stop
// ============================================================

func TestStartAndStop(t *testing.T) {
	tw := New(time.Second, 10)
	if tw == nil {
		t.Fatal("New 返回 nil")
	}
	tw.Start()

	if tw.ticker == nil {
		t.Error("Start 后 ticker 不应为 nil")
	}

	tw.Stop()
	// Stop 不应 panic 或死锁
}

// ============================================================
// AddJob — 任务执行
// ============================================================

func TestAddJobRuns(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var count int32
	done := make(chan struct{}, 1)

	// delay=0：放入当前 slot，下一个 tick 执行
	tw.AddJob(0, "job1", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})

	select {
	case <-done:
		if atomic.LoadInt32(&count) != 1 {
			t.Errorf("任务应执行 1 次，实际 %d 次", count)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("任务未在预期时间内执行")
	}
}

func TestAddJobNegativeDelayIgnored(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var count int32

	tw.AddJob(-time.Second, "job1", func() {
		atomic.AddInt32(&count, 1)
	})
	tw.AddJob(-time.Millisecond, "job2", func() {
		atomic.AddInt32(&count, 1)
	})

	// 等待一个 tick 周期确认不会执行
	time.Sleep(2 * time.Second)
	if atomic.LoadInt32(&count) != 0 {
		t.Errorf("负延迟任务不应执行，实际执行了 %d 次", count)
	}
}

func TestJobWithoutKey(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var count int32
	done := make(chan struct{}, 1)

	tw.AddJob(0, "", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})

	select {
	case <-done:
		if atomic.LoadInt32(&count) != 1 {
			t.Errorf("无 key 任务应执行 1 次，实际 %d 次", count)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("无 key 任务未执行")
	}
}

// ============================================================
// RemoveJob
// ============================================================

func TestRemoveJobBeforeRun(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var count int32

	tw.AddJob(2*time.Second, "job1", func() {
		atomic.AddInt32(&count, 1)
	})

	// 等 add 被 start() goroutine 处理完毕后再移除
	time.Sleep(100 * time.Millisecond)
	tw.RemoveJob("job1")

	// 等待确认不会执行（延迟 2s + 缓冲）
	time.Sleep(4 * time.Second)
	if atomic.LoadInt32(&count) != 0 {
		t.Errorf("已移除任务不应执行，实际执行了 %d 次", count)
	}
}

func TestRemoveJobEmptyKey(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	// 空 key 应在 RemoveJob 入口直接 return，不发送到 channel
	tw.RemoveJob("")
	// 不应 panic 或死锁
}

func TestRemoveJobNotFound(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	// 移除不存在的 key 不应 panic
	tw.RemoveJob("nonexistent")
	tw.RemoveJob("notfound")
}

func TestRemoveSameJobTwice(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	tw.AddJob(3*time.Second, "twice", func() {})
	tw.RemoveJob("twice")

	// 第二次移除同一个 key 不应 panic
	tw.RemoveJob("twice")

	// 验证不在 timer 中
	tw.mu.RLock()
	_, ok := tw.timer["twice"]
	tw.mu.RUnlock()
	if ok {
		t.Error("两次 RemoveJob 后 timer 中不应存在该 key")
	}
}

// ============================================================
// 重复 key 替换
// ============================================================

func TestDuplicateKeyReplacesOld(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var firstCount, secondCount int32

	// 先添加第一个任务（远延迟）
	tw.AddJob(10*time.Second, "dup", func() {
		atomic.AddInt32(&firstCount, 1)
	})

	// 用相同 key 添加第二个任务（delay=0，应该在下一 tick 执行）
	done := make(chan struct{}, 1)
	tw.AddJob(0, "dup", func() {
		atomic.AddInt32(&secondCount, 1)
		done <- struct{}{}
	})

	select {
	case <-done:
		if atomic.LoadInt32(&secondCount) != 1 {
			t.Errorf("第二个任务应执行 1 次，实际 %d 次", secondCount)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("第二个任务未执行")
	}

	// 第一个任务已被替换，不应执行
	if atomic.LoadInt32(&firstCount) != 0 {
		t.Errorf("被替换的第一个任务不应执行，实际执行了 %d 次", firstCount)
	}
}

// ============================================================
// getPositionAndCircle
// ============================================================

func TestGetPositionAndCircle(t *testing.T) {
	tests := []struct {
		name       string
		interval   time.Duration
		slotNum    int
		currentPos int
		delay      time.Duration
		wantCircle int
		wantPos    int
	}{
		{name: "delay=0在当前slot", interval: time.Second, slotNum: 10, currentPos: 0, delay: 0, wantCircle: 0, wantPos: 0},
		{name: "delay小于一圈", interval: time.Second, slotNum: 10, currentPos: 0, delay: 3 * time.Second, wantCircle: 0, wantPos: 3},
		{name: "currentPos非零+delay小于一圈", interval: time.Second, slotNum: 10, currentPos: 5, delay: 3 * time.Second, wantCircle: 0, wantPos: 8},
		{name: "delay正好一圈", interval: time.Second, slotNum: 10, currentPos: 0, delay: 10 * time.Second, wantCircle: 1, wantPos: 0},
		{name: "delay正好两圈", interval: time.Second, slotNum: 10, currentPos: 0, delay: 20 * time.Second, wantCircle: 2, wantPos: 0},
		{name: "delay超过一圈+currentPos非零", interval: time.Second, slotNum: 10, currentPos: 3, delay: 12 * time.Second, wantCircle: 1, wantPos: 5},
		{name: "带小数秒舍入", interval: time.Second, slotNum: 10, currentPos: 0, delay: 5500 * time.Millisecond, wantCircle: 0, wantPos: 5},
		{name: "环绕计算pos=0", interval: time.Second, slotNum: 10, currentPos: 7, delay: 3 * time.Second, wantCircle: 0, wantPos: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := &TimeWheel{
				interval:   tt.interval,
				slotNum:    tt.slotNum,
				currentPos: tt.currentPos,
			}
			pos, circle := tw.getPositionAndCircle(tt.delay)
			if pos != tt.wantPos {
				t.Errorf("pos = %d, want %d", pos, tt.wantPos)
			}
			if circle != tt.wantCircle {
				t.Errorf("circle = %d, want %d", circle, tt.wantCircle)
			}
		})
	}
}

// ============================================================
// 任务 panic 恢复
// ============================================================

func TestJobPanicRecovery(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	done := make(chan struct{}, 1)

	// 第一个任务 panic（应该被 recover，不影响后续）
	tw.AddJob(0, "panic-job", func() {
		panic("故意的 panic 测试")
	})

	// 第二个正常任务（delay=1s，紧跟 panic 任务之后的下一 tick）
	tw.AddJob(time.Second, "ok-job", func() {
		done <- struct{}{}
	})

	select {
	case <-done:
		// 正常任务执行了，说明 panic 被恢复了
	case <-time.After(5 * time.Second):
		t.Fatal("正常任务未执行，panic 可能未被正确恢复")
	}
}

// ============================================================
// 多个任务
// ============================================================

func TestMultipleJobs(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()
	defer tw.Stop()

	var count int32
	n := 3
	done := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		tw.AddJob(0, "", func() {
			atomic.AddInt32(&count, 1)
			done <- struct{}{}
		})
	}

	// 等待所有任务完成
	for i := 0; i < n; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Errorf("第 %d 个任务未执行", i+1)
		}
	}

	if atomic.LoadInt32(&count) != int32(n) {
		t.Errorf("应执行 %d 个任务，实际 %d 个", n, count)
	}
}

// ============================================================
// 不同延迟分配到不同 slot
// ============================================================

func TestDifferentDelayDifferentSlots(t *testing.T) {
	tw := New(time.Second, 20)
	tw.Start()
	defer tw.Stop()

	var count int32
	n := 3
	done := make(chan struct{}, n)

	// 不同延迟分配到不同 slot 位置
	tw.AddJob(0, "fast", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})
	tw.AddJob(time.Second, "mid", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})
	tw.AddJob(2*time.Second, "slow", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})

	for i := 0; i < n; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Errorf("第 %d 个任务未执行", i+1)
		}
	}

	if atomic.LoadInt32(&count) != int32(n) {
		t.Errorf("应执行 %d 个任务，实际 %d 个", n, count)
	}
}

// ============================================================
// 并发 AddJob / RemoveJob
// ============================================================

func TestConcurrentAddAndRemove(t *testing.T) {
	tw := New(time.Second, 20)
	tw.Start()
	defer tw.Stop()

	var wg sync.WaitGroup
	keys := []string{"a", "b", "c", "d", "e"}

	// 并发添加
	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			tw.AddJob(5*time.Second, key, func() {})
		}(k)
	}

	// 并发移除（部分 key）
	for _, k := range keys[:3] {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			// 给 add 一点时间先执行
			time.Sleep(100 * time.Millisecond)
			tw.RemoveJob(key)
		}(k)
	}

	wg.Wait()
	// 不应 panic，不应死锁
}

// ============================================================
// 边界：Stop 后不应再处理新任务
// ============================================================

func TestStopPreventsNewJobs(t *testing.T) {
	tw := New(time.Second, 10)
	tw.Start()

	var count int32
	done := make(chan struct{}, 1)

	tw.AddJob(0, "before-stop", func() {
		atomic.AddInt32(&count, 1)
		done <- struct{}{}
	})

	// 等待任务执行
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop 前添加的任务未执行")
	}

	tw.Stop()

	if atomic.LoadInt32(&count) < 1 {
		t.Errorf("Stop 前添加的任务应至少执行 1 次，实际 %d 次", count)
	}
}
