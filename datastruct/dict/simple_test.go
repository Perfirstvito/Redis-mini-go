package dict

import (
	"sort"
	"testing"
)

// ============================================================
// MakeSimple
// ============================================================

func TestMakeSimple(t *testing.T) {
	d := MakeSimple()
	if d == nil {
		t.Fatal("MakeSimple 返回 nil")
	}
	if d.Len() != 0 {
		t.Errorf("新 dict 长度应为 0，实际 %d", d.Len())
	}
}

// ============================================================
// Put / Get / Len
// ============================================================

func TestSimplePutAndGet(t *testing.T) {
	d := MakeSimple()

	// 插入新 key
	if d.Put("a", 100) != 1 {
		t.Error("Put 新 key 应返回 1")
	}
	if d.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", d.Len())
	}

	// 更新已存在 key
	if d.Put("a", 200) != 0 {
		t.Error("Put 已存在 key 应返回 0")
	}
	if d.Len() != 1 {
		t.Errorf("更新不应改变长度，实际 %d", d.Len())
	}

	// Get
	val, ok := d.Get("a")
	if !ok {
		t.Error("Get 已存在 key 应返回 exists=true")
	}
	if val != 200 {
		t.Errorf("Get 应返回最新值 200，实际 %v", val)
	}

	// Get 不存在的 key
	val, ok = d.Get("not_exist")
	if ok {
		t.Error("Get 不存在的 key 应返回 exists=false")
	}
	if val != nil {
		t.Errorf("Get 不存在 key 的 val 应为 nil，实际 %v", val)
	}
}

func TestSimplePutMultipleKeys(t *testing.T) {
	d := MakeSimple()
	keys := []string{"k1", "k2", "k3", "k4", "k5"}

	for i, k := range keys {
		if d.Put(k, i) != 1 {
			t.Errorf("Put(%q) 应返回 1", k)
		}
	}
	if d.Len() != len(keys) {
		t.Errorf("Len 应为 %d，实际 %d", len(keys), d.Len())
	}
	for i, k := range keys {
		val, ok := d.Get(k)
		if !ok {
			t.Errorf("Get(%q) 应 exists=true", k)
		}
		if val != i {
			t.Errorf("Get(%q) = %v, want %d", k, val, i)
		}
	}
}

// ============================================================
// PutIfAbsent
// ============================================================

func TestSimplePutIfAbsent(t *testing.T) {
	d := MakeSimple()

	// key 不存在时插入
	if d.PutIfAbsent("a", 1) != 1 {
		t.Error("PutIfAbsent key 不存在应返回 1")
	}
	val, _ := d.Get("a")
	if val != 1 {
		t.Errorf("PutIfAbsent 后 Get = %v, want 1", val)
	}

	// key 存在时不更新
	if d.PutIfAbsent("a", 999) != 0 {
		t.Error("PutIfAbsent key 已存在应返回 0")
	}
	val, _ = d.Get("a")
	if val != 1 {
		t.Errorf("PutIfAbsent key 已存在不应覆盖，实际 %v", val)
	}

	if d.Len() != 1 {
		t.Errorf("PutIfAbsent 不改变数量，实际 %d", d.Len())
	}
}

// ============================================================
// PutIfExists
// ============================================================

func TestSimplePutIfExists(t *testing.T) {
	d := MakeSimple()

	// key 不存在时不做任何事
	if d.PutIfExists("a", 1) != 0 {
		t.Error("PutIfExists key 不存在应返回 0")
	}
	if d.Len() != 0 {
		t.Errorf("PutIfExists key 不存在不应添加，实际 Len=%d", d.Len())
	}

	// 先插入再更新
	d.Put("a", 1)
	if d.PutIfExists("a", 999) != 1 {
		t.Error("PutIfExists key 已存在应返回 1")
	}
	val, _ := d.Get("a")
	if val != 999 {
		t.Errorf("PutIfExists 应更新值，实际 %v", val)
	}
	if d.Len() != 1 {
		t.Errorf("PutIfExists 不改变数量，实际 %d", d.Len())
	}
}

// ============================================================
// Remove
// ============================================================

func TestSimpleRemove(t *testing.T) {
	d := MakeSimple()

	// 删除不存在的 key
	val, result := d.Remove("not_exist")
	if result != 0 {
		t.Error("Remove 不存在的 key 应返回 0")
	}
	if val != nil {
		t.Errorf("Remove 不存在的 key val 应为 nil，实际 %v", val)
	}

	// 删除存在的 key
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

	// 再次删除同一个 key
	_, result = d.Remove("a")
	if result != 0 {
		t.Error("Remove 已删除的 key 应返回 0")
	}
}

// ============================================================
// Keys
// ============================================================

func TestSimpleKeys(t *testing.T) {
	d := MakeSimple()

	// 空 dict
	keys := d.Keys()
	if len(keys) != 0 {
		t.Errorf("空 dict Keys 长度应为 0，实际 %d", len(keys))
	}

	// 有数据
	d.Put("b", 1)
	d.Put("a", 2)
	d.Put("c", 3)
	keys = d.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys 长度应为 3，实际 %d", len(keys))
	}

	// 验证包含所有 key（顺序不保证）
	sort.Strings(keys)
	expected := []string{"a", "b", "c"}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("Keys[%d] = %q, want %q", i, keys[i], k)
		}
	}
}

// ============================================================
// ForEach
// ============================================================

func TestSimpleForEach(t *testing.T) {
	d := MakeSimple()
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

func TestSimpleForEachEarlyBreak(t *testing.T) {
	d := MakeSimple()
	d.Put("a", 1)
	d.Put("b", 2)
	d.Put("c", 3)

	count := 0
	d.ForEach(func(key string, val interface{}) bool {
		count++
		return false // 第一次就中断
	})

	if count != 1 {
		t.Errorf("ForEach 中断时应只遍历 1 次，实际 %d", count)
	}
}

// ============================================================
// RandomKeys / RandomDistinctKeys
// ============================================================

func TestSimpleRandomKeys(t *testing.T) {
	d := MakeSimple()
	d.Put("a", 1)
	d.Put("b", 2)

	keys := d.RandomKeys(0)
	if len(keys) != 0 {
		t.Errorf("RandomKeys(0) 长度应为 0，实际 %d", len(keys))
	}

	keys = d.RandomKeys(5)
	if len(keys) != 5 {
		t.Errorf("RandomKeys(5) 长度应为 5，实际 %d", len(keys))
	}
	// 随机 key 应该来自 {"a", "b"}
	for _, k := range keys {
		if k != "a" && k != "b" {
			t.Errorf("RandomKeys 返回了未知 key: %q", k)
		}
	}
}

func TestSimpleRandomDistinctKeys(t *testing.T) {
	d := MakeSimple()
	d.Put("a", 1)
	d.Put("b", 2)

	keys := d.RandomDistinctKeys(0)
	if len(keys) != 0 {
		t.Errorf("RandomDistinctKeys(0) 长度应为 0，实际 %d", len(keys))
	}

	// limit 超过 dict 大小时截断
	keys = d.RandomDistinctKeys(10)
	if len(keys) != 2 {
		t.Errorf("RandomDistinctKeys(10) 应截断为 2，实际 %d", len(keys))
	}

	// 正常情况
	keys = d.RandomDistinctKeys(1)
	if len(keys) != 1 {
		t.Errorf("RandomDistinctKeys(1) 长度应为 1，实际 %d", len(keys))
	}
}

// ============================================================
// Clear
// ============================================================

func TestSimpleClear(t *testing.T) {
	d := MakeSimple()
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
}

// ============================================================
// DictScan
// ============================================================

func TestSimpleDictScanAll(t *testing.T) {
	d := MakeSimple()
	d.Put("key1", []byte("val1"))
	d.Put("key2", []byte("val2"))
	d.Put("key3", []byte("val3"))

	result, nextCursor := d.DictScan(0, 10, "*")
	if nextCursor != 0 {
		t.Errorf("SimpleDict DictScan 游标应始终为 0，实际 %d", nextCursor)
	}
	// 3 keys * 2 (key+value) = 6 entries
	if len(result) != 6 {
		t.Errorf("DictScan * 应返回 6 个元素，实际 %d — %v", len(result), result)
	}
}

func TestSimpleDictScanPattern(t *testing.T) {
	d := MakeSimple()
	d.Put("user:1", []byte("alice"))
	d.Put("user:2", []byte("bob"))
	d.Put("order:1", []byte("item"))
	d.Put("order:2", []byte("item2"))

	// 匹配 user:*
	result, _ := d.DictScan(0, 10, "user:*")
	// 2 keys * 2 = 4
	if len(result) != 4 {
		t.Errorf("DictScan user:* 应返回 4 个元素，实际 %d", len(result))
	}

	// 匹配 order:?
	result, _ = d.DictScan(0, 10, "order:?")
	if len(result) != 4 {
		t.Errorf("DictScan order:? 应返回 4 个元素，实际 %d", len(result))
	}

	// 匹配不存在的 pattern
	result, _ = d.DictScan(0, 10, "notfound:*")
	if len(result) != 0 {
		t.Errorf("DictScan notfound:* 应返回 0 个元素，实际 %d", len(result))
	}
}

func TestSimpleDictScanEmpty(t *testing.T) {
	d := MakeSimple()
	result, _ := d.DictScan(0, 10, "*")
	if len(result) != 0 {
		t.Errorf("空 dict DictScan 应返回空，实际 %d", len(result))
	}
}

func TestSimpleDictScanInvalidPattern(t *testing.T) {
	d := MakeSimple()
	d.Put("key", []byte("val"))
	result, cursor := d.DictScan(0, 10, `[invalid`)
	if cursor != -1 {
		t.Errorf("非法 pattern 游标应为 -1，实际 %d", cursor)
	}
	if len(result) != 0 {
		t.Errorf("非法 pattern 结果应为空，实际 %v", result)
	}
}

// ============================================================
// 边界测试
// ============================================================

func TestSimpleDictVariousTypes(t *testing.T) {
	d := MakeSimple()

	d.Put("int", 42)
	d.Put("string", "hello")
	d.Put("bool", true)
	d.Put("float", 3.14)

	if d.Len() != 4 {
		t.Errorf("Len 应为 4，实际 %d", d.Len())
	}

	intVal, _ := d.Get("int")
	if intVal != 42 {
		t.Errorf("int = %v, want 42", intVal)
	}
	strVal, _ := d.Get("string")
	if strVal != "hello" {
		t.Errorf("string = %v, want hello", strVal)
	}
}
