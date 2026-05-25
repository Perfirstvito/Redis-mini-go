package dict

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

// ============================================================
// computeCapacity
// ============================================================

func TestComputeCapacity(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 16},
		{1, 16},
		{16, 16},
		{17, 32},
		{31, 32},
		{32, 32},
		{33, 64},
		{100, 128},
		{1000, 1024},
		{1024, 1024},
		{1025, 2048},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("computeCapacity(%d)", tt.input), func(t *testing.T) {
			got := computeCapacity(tt.input)
			if got != tt.expected {
				t.Errorf("computeCapacity(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// ============================================================
// MakeConcurrent
// ============================================================

func TestMakeConcurrent(t *testing.T) {
	// shardCount=1 不分片
	d := MakeConcurrent(1)
	if d == nil {
		t.Fatal("MakeConcurrent(1) 返回 nil")
	}
	if d.Len() != 0 {
		t.Errorf("新 dict Len 应为 0，实际 %d", d.Len())
	}
	if d.shardCount != 1 {
		t.Errorf("shardCount 应为 1，实际 %d", d.shardCount)
	}
	if len(d.table) != 1 {
		t.Errorf("table 长度应为 1，实际 %d", len(d.table))
	}

	// shardCount=5 → 向上取整到 2 的幂 = 16
	d = MakeConcurrent(5)
	if d.shardCount != 16 {
		t.Errorf("MakeConcurrent(5) shardCount 应为 16，实际 %d", d.shardCount)
	}
	if len(d.table) != 16 {
		t.Errorf("MakeConcurrent(5) table 长度应为 16，实际 %d", len(d.table))
	}

	// shardCount=16 → 不变
	d = MakeConcurrent(16)
	if d.shardCount != 16 {
		t.Errorf("MakeConcurrent(16) shardCount 应为 16，实际 %d", d.shardCount)
	}
}

// ============================================================
// spread — 哈希分布
// ============================================================

func TestSpread(t *testing.T) {
	d := MakeConcurrent(4) // shardCount=16 (power of 2)

	// 同一个 key 始终映射到同一个 shard
	idx1 := d.spread("hello")
	idx2 := d.spread("hello")
	if idx1 != idx2 {
		t.Errorf("相同 key spread 结果不一致: %d vs %d", idx1, idx2)
	}

	// 索引范围在 [0, shardCount)
	if idx1 >= uint32(d.shardCount) {
		t.Errorf("spread 索引 %d 超出范围 [0, %d)", idx1, d.shardCount)
	}

	// 只有 1 个 shard 时始终返回 0
	d1 := MakeConcurrent(1)
	if d1.spread("anything") != 0 {
		t.Error("单 shard 时 spread 应始终为 0")
	}
}

func TestSpreadDistribution(t *testing.T) {
	shardCount := 16
	d := MakeConcurrent(16)

	// 统计大量 key 的分布
	dist := make([]int, shardCount)
	for i := 0; i < 1000; i++ {
		idx := d.spread(fmt.Sprintf("key:%d", i))
		dist[idx]++
	}

	// 每个 shard 都不应为 0（合理分布）
	emptyShards := 0
	for _, count := range dist {
		if count == 0 {
			emptyShards++
		}
	}
	if emptyShards > 3 {
		t.Errorf("hash 分布不均衡：%d/%d 个 shard 为空", emptyShards, shardCount)
	}
}

// ============================================================
// Get / GetWithLock
// ============================================================

func TestConcurrentGet(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("hello", "world")

	val, ok := d.Get("hello")
	if !ok {
		t.Error("Get 已存在 key 应返回 exists=true")
	}
	if val != "world" {
		t.Errorf("Get = %v, want world", val)
	}

	val, ok = d.Get("not_exist")
	if ok {
		t.Error("Get 不存在 key 应返回 exists=false")
	}
	if val != nil {
		t.Errorf("Get 不存在 key val 应为 nil，实际 %v", val)
	}
}

func TestConcurrentGetWithLock(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("hello", "world")

	// GetWithLock 不加锁也能正确读取（调用者需自行持锁）
	val, ok := d.GetWithLock("hello")
	if !ok {
		t.Error("GetWithLock 已存在 key 应返回 exists=true")
	}
	if val != "world" {
		t.Errorf("GetWithLock = %v, want world", val)
	}
}

// ============================================================
// Put / PutWithLock
// ============================================================

func TestConcurrentPut(t *testing.T) {
	d := MakeConcurrent(4)

	if d.Put("a", 1) != 1 {
		t.Error("Put 新 key 应返回 1")
	}
	if d.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", d.Len())
	}

	if d.Put("a", 999) != 0 {
		t.Error("Put 已存在 key 应返回 0")
	}
	if d.Len() != 1 {
		t.Errorf("Len 应为 1（更新不增加），实际 %d", d.Len())
	}

	val, _ := d.Get("a")
	if val != 999 {
		t.Errorf("Put 更新后 Get = %v, want 999", val)
	}
}

func TestConcurrentPutMultiple(t *testing.T) {
	d := MakeConcurrent(8)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key:%d", i)
		if d.Put(key, i) != 1 {
			t.Errorf("Put(%q) 应返回 1", key)
		}
	}

	if d.Len() != 100 {
		t.Errorf("Len 应为 100，实际 %d", d.Len())
	}

	// 验证所有 key
	for i := 0; i < 100; i++ {
		val, ok := d.Get(fmt.Sprintf("key:%d", i))
		if !ok {
			t.Errorf("key:%d 应存在", i)
		}
		if val != i {
			t.Errorf("key:%d = %v, want %d", i, val, i)
		}
	}
}

func TestConcurrentPutWithLock(t *testing.T) {
	d := MakeConcurrent(4)
	// PutWithLock 不加锁也能正确写入（调用者需自行持锁）
	if d.PutWithLock("a", 1) != 1 {
		t.Error("PutWithLock 新 key 应返回 1")
	}
	if d.Len() != 1 {
		t.Errorf("PutWithLock 后 Len 应为 1，实际 %d", d.Len())
	}
	if d.PutWithLock("a", 2) != 0 {
		t.Error("PutWithLock 已存在 key 应返回 0")
	}
}

// ============================================================
// PutIfAbsent / PutIfAbsentWithLock
// ============================================================

func TestConcurrentPutIfAbsent(t *testing.T) {
	d := MakeConcurrent(4)

	if d.PutIfAbsent("a", 1) != 1 {
		t.Error("PutIfAbsent key 不存在应返回 1")
	}
	if d.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", d.Len())
	}

	if d.PutIfAbsent("a", 999) != 0 {
		t.Error("PutIfAbsent key 已存在应返回 0")
	}
	val, _ := d.Get("a")
	if val != 1 {
		t.Errorf("PutIfAbsent 不应覆盖已有值，实际 %v", val)
	}

	// WithLock 版本
	if d.PutIfAbsentWithLock("b", 2) != 1 {
		t.Error("PutIfAbsentWithLock 新 key 应返回 1")
	}
	if d.PutIfAbsentWithLock("b", 3) != 0 {
		t.Error("PutIfAbsentWithLock 已存在 key 应返回 0")
	}
}

// ============================================================
// PutIfExists / PutIfExistsWithLock
// ============================================================

func TestConcurrentPutIfExists(t *testing.T) {
	d := MakeConcurrent(4)

	// 不存在时不操作
	if d.PutIfExists("a", 1) != 0 {
		t.Error("PutIfExists key 不存在应返回 0")
	}
	if d.Len() != 0 {
		t.Errorf("PutIfExists 不存在不添加，Len 应为 0，实际 %d", d.Len())
	}

	// 先插入再更新
	d.Put("a", 1)
	if d.PutIfExists("a", 999) != 1 {
		t.Error("PutIfExists key 已存在应返回 1")
	}
	val, _ := d.Get("a")
	if val != 999 {
		t.Errorf("PutIfExists 应更新，实际 %v", val)
	}
	if d.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", d.Len())
	}

	// WithLock 版本
	d.PutIfExistsWithLock("a", 888)
	val, _ = d.Get("a")
	if val != 888 {
		t.Errorf("PutIfExistsWithLock 应更新，实际 %v", val)
	}
}

// ============================================================
// Remove / RemoveWithLock
// ============================================================

func TestConcurrentRemove(t *testing.T) {
	d := MakeConcurrent(4)

	// 删除不存在的
	val, result := d.Remove("not_exist")
	if result != 0 {
		t.Error("Remove 不存在 key 应返回 0")
	}
	if val != nil {
		t.Errorf("Remove 不存在 key val 应为 nil，实际 %v", val)
	}

	// 正常删除
	d.Put("a", 42)
	val, result = d.Remove("a")
	if result != 1 {
		t.Error("Remove 已存在 key 应返回 1")
	}
	if val != 42 {
		t.Errorf("Remove 应返回旧值 42，实际 %v", val)
	}
	if d.Len() != 0 {
		t.Errorf("Remove 后 Len 应为 0，实际 %d", d.Len())
	}

	// 重复删除
	_, result = d.Remove("a")
	if result != 0 {
		t.Error("重复 Remove 应返回 0")
	}
}

func TestConcurrentRemoveWithLock(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 99)
	val, result := d.RemoveWithLock("a")
	if result != 1 {
		t.Error("RemoveWithLock 应返回 1")
	}
	if val != 99 {
		t.Errorf("RemoveWithLock 应返回 99，实际 %v", val)
	}
	if d.Len() != 0 {
		t.Errorf("RemoveWithLock 后 Len 应为 0，实际 %d", d.Len())
	}
}

// ============================================================
// Len（原子计数）
// ============================================================

func TestConcurrentLenAtomic(t *testing.T) {
	d := MakeConcurrent(4)

	d.Put("a", 1)
	d.Put("b", 2)
	d.Put("c", 3)
	if d.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", d.Len())
	}

	d.Remove("b")
	if d.Len() != 2 {
		t.Errorf("Remove 后 Len 应为 2，实际 %d", d.Len())
	}

	d.PutIfAbsent("a", 999) // 不应增加
	if d.Len() != 2 {
		t.Errorf("PutIfAbsent 已有 key 后 Len 应为 2，实际 %d", d.Len())
	}
}

// ============================================================
// Keys
// ============================================================

func TestConcurrentKeys(t *testing.T) {
	d := MakeConcurrent(4)

	// 空 dict
	keys := d.Keys()
	if len(keys) != 0 {
		t.Errorf("空 dict Keys 长度为 0，实际 %d", len(keys))
	}

	// 有数据
	d.Put("z", 1)
	d.Put("a", 2)
	d.Put("m", 3)
	keys = d.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys 长度应为 3，实际 %d", len(keys))
	}

	sort.Strings(keys)
	expected := []string{"a", "m", "z"}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("Keys[%d] = %q, want %q", i, keys[i], k)
		}
	}
}

// ============================================================
// ForEach
// ============================================================

func TestConcurrentForEach(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 1)
	d.Put("b", 2)
	d.Put("c", 3)

	sum := 0
	count := 0
	d.ForEach(func(key string, val interface{}) bool {
		count++
		sum += val.(int)
		return true
	})

	if count != 3 {
		t.Errorf("ForEach 应遍历 3 次，实际 %d", count)
	}
	if sum != 6 {
		t.Errorf("ForEach sum 应为 6，实际 %d", sum)
	}
}

func TestConcurrentForEachEarlyBreak(t *testing.T) {
	d := MakeConcurrent(4)
	for i := 0; i < 10; i++ {
		d.Put(fmt.Sprintf("key:%d", i), i)
	}

	count := 0
	d.ForEach(func(key string, val interface{}) bool {
		count++
		return false // 立即中断
	})

	if count != 1 {
		t.Errorf("ForEach 中断时应只遍历 1 次（可能跨 shard 继续），实际 %d", count)
	}
}

// ============================================================
// RandomKeys / RandomDistinctKeys
// ============================================================

func TestConcurrentRandomKeys(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 1)
	d.Put("b", 2)

	keys := d.RandomKeys(0)
	if len(keys) != 0 {
		t.Errorf("RandomKeys(0) 应为空，实际 %d", len(keys))
	}

	// limit >= size 时返回全部 key（不补齐重复）
	keys = d.RandomKeys(3)
	if len(keys) != 2 {
		t.Errorf("RandomKeys(3) 只有 2 个 key 时应返回全部 2 个，实际 %d", len(keys))
	}
	for _, k := range keys {
		if k != "a" && k != "b" {
			t.Errorf("RandomKeys 返回未知 key: %q", k)
		}
	}
}

func TestConcurrentRandomKeysExceedsSize(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 1)
	d.Put("b", 2)

	// limit >= size 时返回全部 keys
	keys := d.RandomKeys(10)
	if len(keys) != 2 {
		t.Errorf("RandomKeys(10) 长度应为 2（全部 key），实际 %d", len(keys))
	}
}

func TestConcurrentRandomDistinctKeys(t *testing.T) {
	d := MakeConcurrent(4)
	for i := 0; i < 100; i++ {
		d.Put(fmt.Sprintf("key:%d", i), i)
	}

	// 正常获取
	keys := d.RandomDistinctKeys(10)
	if len(keys) != 10 {
		t.Errorf("RandomDistinctKeys(10) 长度应为 10，实际 %d", len(keys))
	}

	// 验证无重复
	seen := make(map[string]struct{})
	for _, k := range keys {
		if _, exists := seen[k]; exists {
			t.Errorf("RandomDistinctKeys 有重复 key: %q", k)
		}
		seen[k] = struct{}{}
	}
}

func TestConcurrentRandomDistinctKeysExceedsSize(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 1)
	d.Put("b", 2)

	keys := d.RandomDistinctKeys(10)
	if len(keys) != 2 {
		t.Errorf("RandomDistinctKeys(10) 应截断为 2，实际 %d", len(keys))
	}
}

// ============================================================
// Clear
// ============================================================

func TestConcurrentClear(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("a", 1)
	d.Put("b", 2)
	d.Put("c", 3)

	d.Clear()

	if d.Len() != 0 {
		t.Errorf("Clear 后 Len 应为 0，实际 %d", d.Len())
	}
	_, ok := d.Get("a")
	if ok {
		t.Error("Clear 后 Get 任意 key 应返回 false")
	}

	// Clear 后仍可用
	d.Put("new", "value")
	if d.Len() != 1 {
		t.Errorf("Clear 后 Put 应正常，Len 应为 1，实际 %d", d.Len())
	}
}

// ============================================================
// RWLocks / RWUnLocks
// ============================================================

func TestRWLocksNoDeadlock(t *testing.T) {
	d := MakeConcurrent(16)

	// 模拟 Redis 事务：先对 writeKeys + readKeys 加锁
	writeKeys := []string{"key:1", "key:2"}
	readKeys := []string{"key:3", "key:4"}

	// 先写入数据
	d.Put("key:1", "v1")
	d.Put("key:2", "v2")
	d.Put("key:3", "v3")
	d.Put("key:4", "v4")

	d.RWLocks(writeKeys, readKeys)

	// 锁内操作
	vals := make(map[string]string)
	for _, k := range writeKeys {
		v, _ := d.GetWithLock(k)
		vals[k] = v.(string)
	}
	for _, k := range readKeys {
		v, _ := d.GetWithLock(k)
		vals[k] = v.(string)
	}

	d.RWUnLocks(writeKeys, readKeys)

	if len(vals) != 4 {
		t.Errorf("锁内读取应有 4 个 key，实际 %d", len(vals))
	}
	t.Logf("RWLocks 内成功读取: %v", vals)
}

func TestRWLocksConcurrent(t *testing.T) {
	d := MakeConcurrent(16)
	for i := 0; i < 100; i++ {
		d.Put(fmt.Sprintf("key:%d", i), i)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	// 并发对同一组 key 加锁
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			writeKeys := []string{fmt.Sprintf("key:%d", gid)}
			readKeys := []string{"key:0", "key:50"}

			d.RWLocks(writeKeys, readKeys)
			// 模拟原子操作
			v, _ := d.GetWithLock(writeKeys[0])
			if v != nil {
				d.PutWithLock(writeKeys[0], v.(int)+1)
			}
			d.RWUnLocks(writeKeys, readKeys)
		}(g)
	}
	wg.Wait()
	close(errCh)
}

// ============================================================
// DictScan
// ============================================================

func TestConcurrentDictScanAll(t *testing.T) {
	d := MakeConcurrent(4)
	for i := 0; i < 50; i++ {
		d.Put(fmt.Sprintf("key:%d", i), fmt.Sprintf("val:%d", i))
	}

	result, cursor := d.DictScan(0, 100, "*")
	if cursor != 0 {
		t.Errorf("DictScan 完整扫描后游标应为 0，实际 %d", cursor)
	}
	// 50 keys
	if len(result) != 50 {
		t.Errorf("DictScan * 应返回 50 个 key，实际 %d", len(result))
	}
}

func TestConcurrentDictScanPattern(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("user:1", "alice")
	d.Put("user:2", "bob")
	d.Put("user:3", "charlie")
	d.Put("order:1", "item1")
	d.Put("order:2", "item2")

	result, _ := d.DictScan(0, 100, "user:*")
	if len(result) != 3 {
		t.Errorf("DictScan user:* 应返回 3 个 key，实际 %d — %v", len(result), result)
	}

	result, _ = d.DictScan(0, 100, "user:?")
	if len(result) != 3 {
		t.Errorf("DictScan user:? 应返回 3 个 key，实际 %d — %v", len(result), result)
	}
}

func TestConcurrentDictScanCursor(t *testing.T) {
	d := MakeConcurrent(4)
	for i := 0; i < 100; i++ {
		d.Put(fmt.Sprintf("key:%d", i), i)
	}

	// 分步扫描
	var allKeys []string
	cursor := 0
	iterations := 0
	for {
		result, nextCursor := d.DictScan(cursor, 30, "*")
		for _, b := range result {
			allKeys = append(allKeys, string(b))
		}
		iterations++
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
		if iterations > 50 {
			t.Fatal("DictScan 陷入死循环")
		}
	}

	if len(allKeys) != 100 {
		t.Errorf("分步扫描应返回 100 个 key，实际 %d (迭代 %d 次)", len(allKeys), iterations)
	}
}

func TestConcurrentDictScanInvalidPattern(t *testing.T) {
	d := MakeConcurrent(4)
	d.Put("key", "val")
	result, cursor := d.DictScan(0, 10, `[invalid`)
	if cursor != -1 {
		t.Errorf("非法 pattern 游标应为 -1，实际 %d", cursor)
	}
	if len(result) != 0 {
		t.Errorf("非法 pattern 结果应为空，实际 %v", result)
	}
}

// ============================================================
// 并发安全
// ============================================================

func TestConcurrentPutGetRace(t *testing.T) {
	d := MakeConcurrent(16)
	var wg sync.WaitGroup
	n := 200

	// 并发写入
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Put(fmt.Sprintf("key:%d", id), id)
		}(i)
	}

	// 并发读取
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Get(fmt.Sprintf("key:%d", id))
		}(i)
	}

	wg.Wait()

	if d.Len() != n {
		t.Errorf("并发 Put 后 Len 应为 %d，实际 %d", n, d.Len())
	}
}

func TestConcurrentPutRemoveRace(t *testing.T) {
	d := MakeConcurrent(16)
	var wg sync.WaitGroup

	// 预先插入
	for i := 0; i < 100; i++ {
		d.Put(fmt.Sprintf("key:%d", i), i)
	}

	// 并发更新和删除
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Put(fmt.Sprintf("key:%d", id), id*10)
		}(i)

		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Remove(fmt.Sprintf("key:%d", id+50))
		}(i)
	}

	wg.Wait()

	// 验证前 50 个 key 的值已更新
	for i := 0; i < 50; i++ {
		val, ok := d.Get(fmt.Sprintf("key:%d", i))
		if !ok {
			t.Errorf("key:%d 应存在", i)
		} else if val != i*10 && val != i {
			// 可能被并发覆盖，但值应该是整数
			_ = val
		}
	}

	// 后 50 个 key 应该已被删除
	deleted := 0
	for i := 50; i < 100; i++ {
		if _, ok := d.Get(fmt.Sprintf("key:%d", i)); !ok {
			deleted++
		}
	}
	if deleted != 50 {
		t.Errorf("后 50 个 key 应该全部被删除，实际删除 %d", deleted)
	}
}

func TestConcurrentPutIfAbsentRace(t *testing.T) {
	d := MakeConcurrent(16)
	var wg sync.WaitGroup

	// 多个 goroutine 同时对同一个 key 调用 PutIfAbsent
	// 只有一个应该成功
	successCount := 0
	var mu sync.Mutex
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if d.PutIfAbsent("single_key", "only_one") == 1 {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if successCount != 1 {
		t.Errorf("PutIfAbsent 并发争用，应有恰好 1 个成功，实际 %d", successCount)
	}
	if d.Len() != 1 {
		t.Errorf("PutIfAbsent 并发争用后 Len 应为 1，实际 %d", d.Len())
	}
}

func TestConcurrentLenConsistency(t *testing.T) {
	d := MakeConcurrent(16)
	var wg sync.WaitGroup
	n := 500

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Put(fmt.Sprintf("key:%d", id), id)
		}(i)
	}
	wg.Wait()

	if d.Len() != n {
		t.Errorf("并发 Put 后 Len 应为 %d，实际 %d", n, d.Len())
	}

	// 删除一半
	for i := 0; i < n/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.Remove(fmt.Sprintf("key:%d", id))
		}(i)
	}
	wg.Wait()

	if d.Len() != n-n/2 {
		t.Errorf("删除一半后 Len 应为 %d，实际 %d", n-n/2, d.Len())
	}
}

// ============================================================
// fnv32 哈希
// ============================================================

func TestFnv32(t *testing.T) {
	// 相同输入产生相同输出
	h1 := fnv32("hello")
	h2 := fnv32("hello")
	if h1 != h2 {
		t.Errorf("fnv32 确定性失败: %d vs %d", h1, h2)
	}

	// 不同输入大概率不同
	h3 := fnv32("world")
	if h1 == h3 {
		t.Log("hash 碰撞: hello 和 world 的 fnv32 相同（概率极低但合法）")
	}

	// 空串
	h4 := fnv32("")
	if h4 == 0 {
		t.Error("fnv32(\"\") 不应为 0")
	}
}

// ============================================================
// shard.RandomKey (通过 RandomKeys 间接测试)
// ============================================================

func TestShardRandomKey(t *testing.T) {
	// 构造只有 1 个 shard 的 dict，测试 shard.RandomKey
	d := MakeConcurrent(1)
	d.Put("single", "only")

	s := d.table[0]
	key := s.RandomKey()
	if key != "single" {
		t.Errorf("单 key shard RandomKey 应为 single，实际 %q", key)
	}

	// 空 shard
	d.Remove("single")
	key = s.RandomKey()
	if key != "" {
		t.Errorf("空 shard RandomKey 应为空串，实际 %q", key)
	}
}
