package list

import (
	"testing"
)

// ============================================================
// Make — 构造函数
// ============================================================

func TestMake_Empty(t *testing.T) {
	list := Make()
	if list == nil {
		t.Fatal("Make() 返回 nil")
	}
	if list.Len() != 0 {
		t.Errorf("空列表 Len 应为 0，实际 %d", list.Len())
	}
}

func TestMake_WithValues(t *testing.T) {
	list := Make(1, 2, 3, 4, 5)
	if list.Len() != 5 {
		t.Errorf("Len 应为 5，实际 %d", list.Len())
	}
	for i := 0; i < 5; i++ {
		if list.Get(i) != i+1 {
			t.Errorf("Get(%d) = %v, want %d", i, list.Get(i), i+1)
		}
	}
}

// ============================================================
// Add — 尾部追加
// ============================================================

func TestAdd_Single(t *testing.T) {
	list := Make()
	list.Add("hello")
	if list.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", list.Len())
	}
	if list.Get(0) != "hello" {
		t.Errorf("Get(0) = %v, want hello", list.Get(0))
	}
}

func TestAdd_Multiple(t *testing.T) {
	list := Make()
	vals := []interface{}{"a", "b", "c", 42, true}
	for _, v := range vals {
		list.Add(v)
	}
	if list.Len() != len(vals) {
		t.Errorf("Len 应为 %d，实际 %d", len(vals), list.Len())
	}
	for i, v := range vals {
		if list.Get(i) != v {
			t.Errorf("Get(%d) = %v, want %v", i, list.Get(i), v)
		}
	}
}

// ============================================================
// Get — 按索引取值
// ============================================================

func TestGet_FirstLast(t *testing.T) {
	list := Make(1, 2, 3)
	if list.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", list.Get(0))
	}
	if list.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", list.Get(2))
	}
}

func TestGet_Middle(t *testing.T) {
	list := Make(10, 20, 30, 40, 50)
	if list.Get(2) != 30 {
		t.Errorf("Get(2) = %v, want 30", list.Get(2))
	}
}

func TestGet_PanicOnNegative(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get(-1) 应 panic")
		}
	}()
	list.Get(-1)
}

func TestGet_PanicOnOutOfBound(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get(list.Len()) 应 panic")
		}
	}()
	list.Get(list.Len())
}

// ============================================================
// Set — 按索引更新
// ============================================================

func TestSet_ValidIndex(t *testing.T) {
	list := Make(1, 2, 3)
	list.Set(1, 999)
	if list.Get(1) != 999 {
		t.Errorf("Set(1,999) 后 Get(1) = %v, want 999", list.Get(1))
	}
}

func TestSet_FirstAndLast(t *testing.T) {
	list := Make(1, 2, 3)
	list.Set(0, "first")
	list.Set(2, "last")
	if list.Get(0) != "first" {
		t.Errorf("Set(0) = %v, want first", list.Get(0))
	}
	if list.Get(2) != "last" {
		t.Errorf("Set(2) = %v, want last", list.Get(2))
	}
}

func TestSet_PanicOnNegative(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Set(-1) 应 panic")
		}
	}()
	list.Set(-1, "x")
}

// ============================================================
// Insert — 按索引插入（原元素后移）
// ============================================================

func TestInsert_Front(t *testing.T) {
	list := Make(2, 3)
	list.Insert(0, 1)
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	if list.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", list.Get(0))
	}
	if list.Get(1) != 2 {
		t.Errorf("Get(1) = %v, want 2", list.Get(1))
	}
}

func TestInsert_Middle(t *testing.T) {
	list := Make(1, 3)
	list.Insert(1, 2)
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	if list.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", list.Get(0))
	}
	if list.Get(1) != 2 {
		t.Errorf("Get(1) = %v, want 2", list.Get(1))
	}
	if list.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", list.Get(2))
	}
}

func TestInsert_Back_SameAsAdd(t *testing.T) {
	list := Make(1, 2)
	list.Insert(list.Len(), 3) // index == size → append
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	if list.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", list.Get(2))
	}
}

func TestInsert_EmptyList(t *testing.T) {
	list := Make()
	list.Insert(0, "only")
	if list.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", list.Len())
	}
	if list.Get(0) != "only" {
		t.Errorf("Get(0) = %v, want only", list.Get(0))
	}
}

func TestInsert_PanicOnNegative(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Insert(-1) 应 panic")
		}
	}()
	list.Insert(-1, "x")
}

func TestInsert_PanicOnTooLarge(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Insert(size+1) 应 panic")
		}
	}()
	list.Insert(list.Len()+1, "x")
}

// ============================================================
// Remove — 按索引删除
// ============================================================

func TestRemove_Front(t *testing.T) {
	list := Make(1, 2, 3)
	val := list.Remove(0)
	if val != 1 {
		t.Errorf("Remove(0) 应返回 1，实际 %v", val)
	}
	if list.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", list.Len())
	}
	if list.Get(0) != 2 {
		t.Errorf("删除头后 Get(0) = %v, want 2", list.Get(0))
	}
}

func TestRemove_Middle(t *testing.T) {
	list := Make(1, 2, 3)
	val := list.Remove(1)
	if val != 2 {
		t.Errorf("Remove(1) 应返回 2，实际 %v", val)
	}
	if list.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", list.Len())
	}
	if list.Get(0) != 1 || list.Get(1) != 3 {
		t.Errorf("删除中间后应为 [1, 3]，实际 [%v, %v]", list.Get(0), list.Get(1))
	}
}

func TestRemove_Last(t *testing.T) {
	list := Make(1, 2, 3)
	val := list.Remove(2)
	if val != 3 {
		t.Errorf("Remove(2) 应返回 3，实际 %v", val)
	}
	if list.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", list.Len())
	}
}

func TestRemove_OnlyElement(t *testing.T) {
	list := Make(42)
	val := list.Remove(0)
	if val != 42 {
		t.Errorf("Remove(0) 应返回 42，实际 %v", val)
	}
	if list.Len() != 0 {
		t.Errorf("删除唯一元素后 Len 应为 0，实际 %d", list.Len())
	}
	// 验证 first 和 last 都被正确置空（通过 RemoveLast 验证）
	if list.RemoveLast() != nil {
		t.Error("空列表 RemoveLast 应返回 nil")
	}
}

func TestRemove_PanicOnNegative(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Remove(-1) 应 panic")
		}
	}()
	list.Remove(-1)
}

func TestRemove_PanicOnOutOfBound(t *testing.T) {
	list := Make(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Remove(size) 应 panic")
		}
	}()
	list.Remove(list.Len())
}

// ============================================================
// RemoveLast — 删除尾部元素
// ============================================================

func TestRemoveLast_Normal(t *testing.T) {
	list := Make(1, 2, 3)
	val := list.RemoveLast()
	if val != 3 {
		t.Errorf("RemoveLast 应返回 3，实际 %v", val)
	}
	if list.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", list.Len())
	}
	if list.Get(list.Len()-1) != 2 {
		t.Errorf("新的尾部应为 2，实际 %v", list.Get(list.Len()-1))
	}
}

func TestRemoveLast_Single(t *testing.T) {
	list := Make("only")
	val := list.RemoveLast()
	if val != "only" {
		t.Errorf("RemoveLast 应返回 only，实际 %v", val)
	}
	if list.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", list.Len())
	}
}

func TestRemoveLast_Empty(t *testing.T) {
	list := Make()
	val := list.RemoveLast()
	if val != nil {
		t.Errorf("空列表 RemoveLast 应返回 nil，实际 %v", val)
	}
	if list.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", list.Len())
	}
}

func TestRemoveLast_Chain(t *testing.T) {
	list := Make(1, 2, 3, 4, 5)
	for i := 5; i >= 1; i-- {
		val := list.RemoveLast()
		if val != i {
			t.Errorf("RemoveLast #%d = %v, want %d", 6-i, val, i)
		}
	}
	if list.Len() != 0 {
		t.Errorf("链式 RemoveLast 后 Len 应为 0，实际 %d", list.Len())
	}
}

// ============================================================
// RemoveAllByVal — 删除所有匹配值
// ============================================================

func TestRemoveAllByVal_RemoveAll(t *testing.T) {
	list := Make(1, 2, 1, 3, 1, 4)
	removed := list.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 1
	})
	if removed != 3 {
		t.Errorf("应删除 3 个 1，实际 %d", removed)
	}
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	// 验证顺序：2, 3, 4
	if list.Get(0) != 2 || list.Get(1) != 3 || list.Get(2) != 4 {
		t.Errorf("删除后应为 [2, 3, 4]，实际 [%v, %v, %v]",
			list.Get(0), list.Get(1), list.Get(2))
	}
}

func TestRemoveAllByVal_RemoveNone(t *testing.T) {
	list := Make(1, 2, 3)
	removed := list.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 999
	})
	if removed != 0 {
		t.Errorf("应删除 0 个，实际 %d", removed)
	}
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
}

func TestRemoveAllByVal_RemoveAllElements(t *testing.T) {
	list := Make(5, 5, 5)
	removed := list.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 5
	})
	if removed != 3 {
		t.Errorf("应删除 3 个，实际 %d", removed)
	}
	if list.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", list.Len())
	}
	if list.RemoveLast() != nil {
		t.Error("空列表 RemoveLast 应返回 nil")
	}
}

func TestRemoveAllByVal_String(t *testing.T) {
	list := Make("a", "b", "a", "c", "a")
	removed := list.RemoveAllByVal(func(a interface{}) bool {
		return a.(string) == "a"
	})
	if removed != 3 {
		t.Errorf("应删除 3 个 a，实际 %d", removed)
	}
	if list.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", list.Len())
	}
}

// ============================================================
// RemoveByVal — 左到右删除最多 count 个
// ============================================================

func TestRemoveByVal_Partial(t *testing.T) {
	list := Make(1, 1, 1, 2, 3)
	removed := list.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 1，实际 %d", removed)
	}
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	// 还剩 1, 2, 3
	if list.Get(0) != 1 {
		t.Errorf("应保留第三个 1，实际 Get(0)=%v", list.Get(0))
	}
}

func TestRemoveByVal_CountExceeds(t *testing.T) {
	list := Make(1, 1, 2)
	removed := list.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 100) // count > actual
	if removed != 2 {
		t.Errorf("应删除 2 个（不超过实际），实际 %d", removed)
	}
	if list.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", list.Len())
	}
}

func TestRemoveByVal_Zero(t *testing.T) {
	// Redis LREM 语义：count=0 表示删除所有匹配元素
	list := Make(1, 1, 1)
	removed := list.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 0)
	if removed != 3 {
		t.Errorf("count=0（Redis 语义：删除全部）应删除 3 个，实际 %d", removed)
	}
	if list.Len() != 0 {
		t.Errorf("count=0 删除全部后 Len 应为 0，实际 %d", list.Len())
	}
}

// ============================================================
// ReverseRemoveByVal — 右到左删除最多 count 个
// ============================================================

func TestReverseRemoveByVal_Partial(t *testing.T) {
	list := Make(1, 1, 1, 2, 3)
	removed := list.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 1（从右），实际 %d", removed)
	}
	if list.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", list.Len())
	}
	// 从左起只剩第一个 1
	if list.Get(0) != 1 || list.Get(1) != 2 || list.Get(2) != 3 {
		t.Errorf("反向删除后应为 [1, 2, 3]，实际 [%v, %v, %v]",
			list.Get(0), list.Get(1), list.Get(2))
	}
}

func TestReverseRemoveByVal_All(t *testing.T) {
	list := Make("x", "y", "x")
	removed := list.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(string) == "x"
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 x，实际 %d", removed)
	}
	if list.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", list.Len())
	}
}

func TestReverseRemoveByVal_CountExceeds(t *testing.T) {
	list := Make(1, 1)
	removed := list.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 5)
	if removed != 2 {
		t.Errorf("count > 实际时删除全部，实际 %d", removed)
	}
	if list.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", list.Len())
	}
}

// ============================================================
// Len — 长度
// ============================================================

func TestLen_AfterOperations(t *testing.T) {
	list := Make()
	if list.Len() != 0 {
		t.Errorf("初始 Len 应为 0，实际 %d", list.Len())
	}
	list.Add(1)
	list.Add(2)
	if list.Len() != 2 {
		t.Errorf("Add 后 Len 应为 2，实际 %d", list.Len())
	}
	list.Insert(1, "mid")
	if list.Len() != 3 {
		t.Errorf("Insert 后 Len 应为 3，实际 %d", list.Len())
	}
	list.Remove(1)
	if list.Len() != 2 {
		t.Errorf("Remove 后 Len 应为 2，实际 %d", list.Len())
	}
	list.RemoveLast()
	if list.Len() != 1 {
		t.Errorf("RemoveLast 后 Len 应为 1，实际 %d", list.Len())
	}
}

// ============================================================
// ForEach — 遍历
// ============================================================

func TestForEach_FullTraversal(t *testing.T) {
	list := Make(10, 20, 30, 40)
	sum := 0
	count := 0
	list.ForEach(func(i int, v interface{}) bool {
		count++
		sum += v.(int)
		return true
	})
	if count != 4 {
		t.Errorf("ForEach 应遍历 4 个元素，实际 %d", count)
	}
	if sum != 100 {
		t.Errorf("sum 应为 100，实际 %d", sum)
	}
}

func TestForEach_EarlyBreak(t *testing.T) {
	list := Make(1, 2, 3, 4, 5)
	count := 0
	list.ForEach(func(i int, v interface{}) bool {
		count++
		return v.(int) < 3
	})
	if count != 3 {
		t.Errorf("ForEach 应在 v=3 时中断，遍历次数 3，实际 %d", count)
	}
}

func TestForEach_Index(t *testing.T) {
	list := Make("a", "b", "c")
	indices := make([]int, 0)
	list.ForEach(func(i int, v interface{}) bool {
		indices = append(indices, i)
		return true
	})
	expected := []int{0, 1, 2}
	for i, idx := range expected {
		if indices[i] != idx {
			t.Errorf("ForEach index[%d] = %d, want %d", i, indices[i], idx)
		}
	}
}

// ============================================================
// Contains — 判断是否包含
// ============================================================

func TestContains_Found(t *testing.T) {
	list := Make(1, 2, 3, 4, 5)
	if !list.Contains(func(a interface{}) bool {
		return a.(int) == 3
	}) {
		t.Error("Contains(3) 应为 true")
	}
}

func TestContains_NotFound(t *testing.T) {
	list := Make(1, 2, 3)
	if list.Contains(func(a interface{}) bool {
		return a.(int) == 999
	}) {
		t.Error("Contains(999) 应为 false")
	}
}

func TestContains_First(t *testing.T) {
	list := Make(1, 2, 3)
	if !list.Contains(func(a interface{}) bool {
		return a.(int) == 1
	}) {
		t.Error("Contains(1) 第一个元素 应为 true")
	}
}

func TestContains_Last(t *testing.T) {
	list := Make(1, 2, 3)
	if !list.Contains(func(a interface{}) bool {
		return a.(int) == 3
	}) {
		t.Error("Contains(3) 最后一个元素 应为 true")
	}
}

func TestContains_Empty(t *testing.T) {
	list := Make()
	if list.Contains(func(a interface{}) bool {
		return true
	}) {
		t.Error("空列表 Contains 应为 false")
	}
}

// ============================================================
// Range — 切片 [start, stop)
// ============================================================

func TestRange_Normal(t *testing.T) {
	list := Make(1, 2, 3, 4, 5)
	slice := list.Range(1, 4)
	if len(slice) != 3 {
		t.Errorf("Range(1,4) 长度应为 3，实际 %d", len(slice))
	}
	expected := []interface{}{2, 3, 4}
	for i, v := range expected {
		if slice[i] != v {
			t.Errorf("Range[%d] = %v, want %v", i, slice[i], v)
		}
	}
}

func TestRange_Full(t *testing.T) {
	list := Make(1, 2, 3)
	slice := list.Range(0, list.Len())
	if len(slice) != 3 {
		t.Errorf("Range(0,size) 长度应为 3，实际 %d", len(slice))
	}
}

func TestRange_Single(t *testing.T) {
	list := Make(1, 2, 3)
	slice := list.Range(1, 2)
	if len(slice) != 1 || slice[0] != 2 {
		t.Errorf("Range(1,2) = %v, want [2]", slice)
	}
}

func TestRange_StartEqualsStop(t *testing.T) {
	list := Make(1, 2, 3)
	slice := list.Range(1, 1)
	if len(slice) != 0 {
		t.Errorf("Range(1,1) 应为空切片，实际 %v", slice)
	}
}

func TestRange_PanicOnNegativeStart(t *testing.T) {
	list := Make(1, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(-1, 1) 应 panic")
		}
	}()
	list.Range(-1, 1)
}

func TestRange_PanicOnStartOutOfRange(t *testing.T) {
	list := Make(1, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(2, 2) start>=size 应 panic")
		}
	}()
	list.Range(2, 2)
}

func TestRange_PanicOnStopBeforeStart(t *testing.T) {
	list := Make(1, 2, 3)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(2, 1) 应 panic")
		}
	}()
	list.Range(2, 1)
}

func TestRange_PanicOnStopOutOfRange(t *testing.T) {
	list := Make(1, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(0, 3) 应 panic")
		}
	}()
	list.Range(0, 3)
}

// ============================================================
// find — 内部二分查找 (通过 Get 间接测试)
// ============================================================

func TestFind_Optimization(t *testing.T) {
	// 构造大列表验证 find 两端遍历优化
	list := Make()
	for i := 0; i < 100; i++ {
		list.Add(i)
	}

	// 前半部分从 first 走
	if list.Get(10) != 10 {
		t.Errorf("Get(10) = %v, want 10", list.Get(10))
	}
	// 后半部分从 last 走
	if list.Get(90) != 90 {
		t.Errorf("Get(90) = %v, want 90", list.Get(90))
	}
}

// ============================================================
// Panic: nil list
// ============================================================

func TestNilList_Panics(t *testing.T) {
	var list *LinkedList

	// 每个方法在 nil receiver 上调用都应 panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Add", func() { list.Add(1) }},
		{"Get", func() { list.Get(0) }},
		{"Set", func() { list.Set(0, 1) }},
		{"Insert", func() { list.Insert(0, 1) }},
		{"Remove", func() { list.Remove(0) }},
		{"RemoveLast", func() { list.RemoveLast() }},
		{"RemoveAllByVal", func() { list.RemoveAllByVal(func(a interface{}) bool { return true }) }},
		{"RemoveByVal", func() { list.RemoveByVal(func(a interface{}) bool { return true }, 1) }},
		{"ReverseRemoveByVal", func() { list.ReverseRemoveByVal(func(a interface{}) bool { return true }, 1) }},
		{"Len", func() { list.Len() }},
		{"ForEach", func() { list.ForEach(func(i int, v interface{}) bool { return true }) }},
		{"Contains", func() { list.Contains(func(a interface{}) bool { return true }) }},
		{"Range", func() { list.Range(0, 1) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("nil list.%s 应 panic", tt.name)
				}
			}()
			tt.fn()
		})
	}
}

// ============================================================
// 综合场景
// ============================================================

func TestLinkedList_Comprehensive(t *testing.T) {
	list := Make()

	// 构建列表: [10, 20, 30, 40, 50]
	list.Add(10)
	list.Add(20)
	list.Add(30)
	list.Add(40)
	list.Add(50)

	if list.Len() != 5 {
		t.Fatalf("初始 Len 应为 5，实际 %d", list.Len())
	}

	// Insert: [10, 15, 20, 30, 40, 50]
	list.Insert(1, 15)
	if list.Get(1) != 15 {
		t.Errorf("Insert 后 Get(1) = %v, want 15", list.Get(1))
	}

	// Set: [10, 15, 20, 30, 40, 60]
	list.Set(list.Len()-1, 60)
	if list.Get(list.Len()-1) != 60 {
		t.Errorf("Set 后尾部 = %v, want 60", list.Get(list.Len()-1))
	}

	// Remove: [10, 15, 30, 40, 60]
	val := list.Remove(2)
	if val != 20 {
		t.Errorf("Remove(2) = %v, want 20", val)
	}

	// RemoveLast: [10, 15, 30, 40]
	val = list.RemoveLast()
	if val != 60 {
		t.Errorf("RemoveLast = %v, want 60", val)
	}

	// RemoveAllByVal: 删除所有 40 → [10, 15, 30]
	removed := list.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 40
	})
	if removed != 1 {
		t.Errorf("RemoveAllByVal = %d, want 1", removed)
	}

	// 最终验证
	expected := []int{10, 15, 30}
	if list.Len() != len(expected) {
		t.Fatalf("最终 Len 应为 %d，实际 %d", len(expected), list.Len())
	}
	for i, want := range expected {
		if list.Get(i) != want {
			t.Errorf("最终 Get(%d) = %v, want %d", i, list.Get(i), want)
		}
	}
}
