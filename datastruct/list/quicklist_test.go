package list

import (
	"testing"
)

// ============================================================
// NewQuickList — 构造函数
// ============================================================

func TestNewQuickList_Empty(t *testing.T) {
	ql := NewQuickList()
	if ql == nil {
		t.Fatal("NewQuickList() 返回 nil")
	}
	if ql.Len() != 0 {
		t.Errorf("空列表 Len 应为 0，实际 %d", ql.Len())
	}
}

// ============================================================
// Add — 尾部追加
// ============================================================

func TestQuickList_Add_Single(t *testing.T) {
	ql := NewQuickList()
	ql.Add("hello")
	if ql.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", ql.Len())
	}
	if ql.Get(0) != "hello" {
		t.Errorf("Get(0) = %v, want hello", ql.Get(0))
	}
}

func TestQuickList_Add_Multiple(t *testing.T) {
	ql := NewQuickList()
	vals := []interface{}{"a", "b", "c", 42, true}
	for _, v := range vals {
		ql.Add(v)
	}
	if ql.Len() != len(vals) {
		t.Errorf("Len 应为 %d，实际 %d", len(vals), ql.Len())
	}
	for i, v := range vals {
		if ql.Get(i) != v {
			t.Errorf("Get(%d) = %v, want %v", i, ql.Get(i), v)
		}
	}
}

// ============================================================
// Get — 按索引取值
// ============================================================

func TestQuickList_Get_FirstLast(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	if ql.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", ql.Get(0))
	}
	if ql.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", ql.Get(2))
	}
}

func TestQuickList_Get_Middle(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{10, 20, 30, 40, 50} {
		ql.Add(v)
	}
	if ql.Get(2) != 30 {
		t.Errorf("Get(2) = %v, want 30", ql.Get(2))
	}
}

func TestQuickList_Get_PanicOnNegative(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get(-1) 应 panic")
		}
	}()
	ql.Get(-1)
}

func TestQuickList_Get_PanicOnOutOfBound(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get(Len()) 应 panic")
		}
	}()
	ql.Get(ql.Len())
}

// ============================================================
// Set — 按索引更新
// ============================================================

func TestQuickList_Set_ValidIndex(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	ql.Set(1, 999)
	if ql.Get(1) != 999 {
		t.Errorf("Set(1,999) 后 Get(1) = %v, want 999", ql.Get(1))
	}
}

func TestQuickList_Set_FirstAndLast(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	ql.Set(0, "first")
	ql.Set(2, "last")
	if ql.Get(0) != "first" {
		t.Errorf("Set(0) = %v, want first", ql.Get(0))
	}
	if ql.Get(2) != "last" {
		t.Errorf("Set(2) = %v, want last", ql.Get(2))
	}
}

func TestQuickList_Set_PanicOnNegative(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Set(-1) 应 panic")
		}
	}()
	ql.Set(-1, "x")
}

// ============================================================
// Insert — 按索引插入（原元素后移）
// ============================================================

func TestQuickList_Insert_Front(t *testing.T) {
	ql := NewQuickList()
	ql.Add(2)
	ql.Add(3)
	ql.Insert(0, 1)
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	if ql.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", ql.Get(0))
	}
	if ql.Get(1) != 2 {
		t.Errorf("Get(1) = %v, want 2", ql.Get(1))
	}
}

func TestQuickList_Insert_Middle(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(3)
	ql.Insert(1, 2)
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	if ql.Get(0) != 1 {
		t.Errorf("Get(0) = %v, want 1", ql.Get(0))
	}
	if ql.Get(1) != 2 {
		t.Errorf("Get(1) = %v, want 2", ql.Get(1))
	}
	if ql.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", ql.Get(2))
	}
}

func TestQuickList_Insert_Back_SameAsAdd(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Insert(ql.Len(), 3) // index == size → append
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	if ql.Get(2) != 3 {
		t.Errorf("Get(2) = %v, want 3", ql.Get(2))
	}
}

func TestQuickList_Insert_EmptyList(t *testing.T) {
	ql := NewQuickList()
	ql.Insert(0, "only")
	if ql.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", ql.Len())
	}
	if ql.Get(0) != "only" {
		t.Errorf("Get(0) = %v, want only", ql.Get(0))
	}
}

func TestQuickList_Insert_PanicOnNegative(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Insert(-1) 应 panic")
		}
	}()
	ql.Insert(-1, "x")
}

func TestQuickList_Insert_PanicOnTooLarge(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Insert(size+1) 应 panic")
		}
	}()
	ql.Insert(ql.Len()+1, "x")
}

// ============================================================
// Remove — 按索引删除
// ============================================================

func TestQuickList_Remove_Front(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	val := ql.Remove(0)
	if val != 1 {
		t.Errorf("Remove(0) 应返回 1，实际 %v", val)
	}
	if ql.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", ql.Len())
	}
	if ql.Get(0) != 2 {
		t.Errorf("删除头后 Get(0) = %v, want 2", ql.Get(0))
	}
}

func TestQuickList_Remove_Middle(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	val := ql.Remove(1)
	if val != 2 {
		t.Errorf("Remove(1) 应返回 2，实际 %v", val)
	}
	if ql.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", ql.Len())
	}
	if ql.Get(0) != 1 || ql.Get(1) != 3 {
		t.Errorf("删除中间后应为 [1, 3]，实际 [%v, %v]", ql.Get(0), ql.Get(1))
	}
}

func TestQuickList_Remove_Last(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	val := ql.Remove(2)
	if val != 3 {
		t.Errorf("Remove(2) 应返回 3，实际 %v", val)
	}
	if ql.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", ql.Len())
	}
}

func TestQuickList_Remove_OnlyElement(t *testing.T) {
	ql := NewQuickList()
	ql.Add(42)
	val := ql.Remove(0)
	if val != 42 {
		t.Errorf("Remove(0) 应返回 42，实际 %v", val)
	}
	if ql.Len() != 0 {
		t.Errorf("删除唯一元素后 Len 应为 0，实际 %d", ql.Len())
	}
	if ql.RemoveLast() != nil {
		t.Error("空列表 RemoveLast 应返回 nil")
	}
}

func TestQuickList_Remove_PanicOnNegative(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Remove(-1) 应 panic")
		}
	}()
	ql.Remove(-1)
}

func TestQuickList_Remove_PanicOnOutOfBound(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Remove(size) 应 panic")
		}
	}()
	ql.Remove(ql.Len())
}

// ============================================================
// RemoveLast — 删除尾部元素
// ============================================================

func TestQuickList_RemoveLast_Normal(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	val := ql.RemoveLast()
	if val != 3 {
		t.Errorf("RemoveLast 应返回 3，实际 %v", val)
	}
	if ql.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", ql.Len())
	}
	if ql.Get(ql.Len()-1) != 2 {
		t.Errorf("新的尾部应为 2，实际 %v", ql.Get(ql.Len()-1))
	}
}

func TestQuickList_RemoveLast_Single(t *testing.T) {
	ql := NewQuickList()
	ql.Add("only")
	val := ql.RemoveLast()
	if val != "only" {
		t.Errorf("RemoveLast 应返回 only，实际 %v", val)
	}
	if ql.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", ql.Len())
	}
}

func TestQuickList_RemoveLast_Empty(t *testing.T) {
	ql := NewQuickList()
	val := ql.RemoveLast()
	if val != nil {
		t.Errorf("空列表 RemoveLast 应返回 nil，实际 %v", val)
	}
	if ql.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", ql.Len())
	}
}

func TestQuickList_RemoveLast_Chain(t *testing.T) {
	ql := NewQuickList()
	for i := 1; i <= 5; i++ {
		ql.Add(i)
	}
	for i := 5; i >= 1; i-- {
		val := ql.RemoveLast()
		if val != i {
			t.Errorf("RemoveLast #%d = %v, want %d", 6-i, val, i)
		}
	}
	if ql.Len() != 0 {
		t.Errorf("链式 RemoveLast 后 Len 应为 0，实际 %d", ql.Len())
	}
}

// ============================================================
// RemoveAllByVal — 删除所有匹配值
// ============================================================

func TestQuickList_RemoveAllByVal_RemoveAll(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 2, 1, 3, 1, 4} {
		ql.Add(v)
	}
	removed := ql.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 1
	})
	if removed != 3 {
		t.Errorf("应删除 3 个 1，实际 %d", removed)
	}
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	if ql.Get(0) != 2 || ql.Get(1) != 3 || ql.Get(2) != 4 {
		t.Errorf("删除后应为 [2, 3, 4]，实际 [%v, %v, %v]",
			ql.Get(0), ql.Get(1), ql.Get(2))
	}
}

func TestQuickList_RemoveAllByVal_RemoveNone(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	removed := ql.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 999
	})
	if removed != 0 {
		t.Errorf("应删除 0 个，实际 %d", removed)
	}
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
}

func TestQuickList_RemoveAllByVal_RemoveAllElements(t *testing.T) {
	ql := NewQuickList()
	ql.Add(5)
	ql.Add(5)
	ql.Add(5)
	removed := ql.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 5
	})
	if removed != 3 {
		t.Errorf("应删除 3 个，实际 %d", removed)
	}
	if ql.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", ql.Len())
	}
	if ql.RemoveLast() != nil {
		t.Error("空列表 RemoveLast 应返回 nil")
	}
}

func TestQuickList_RemoveAllByVal_String(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []string{"a", "b", "a", "c", "a"} {
		ql.Add(v)
	}
	removed := ql.RemoveAllByVal(func(a interface{}) bool {
		return a.(string) == "a"
	})
	if removed != 3 {
		t.Errorf("应删除 3 个 a，实际 %d", removed)
	}
	if ql.Len() != 2 {
		t.Errorf("Len 应为 2，实际 %d", ql.Len())
	}
}

// ============================================================
// RemoveByVal — 左到右删除最多 count 个
// ============================================================

func TestQuickList_RemoveByVal_Partial(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 1, 1, 2, 3} {
		ql.Add(v)
	}
	removed := ql.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 1，实际 %d", removed)
	}
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	// 还剩 1, 2, 3
	if ql.Get(0) != 1 {
		t.Errorf("应保留第三个 1，实际 Get(0)=%v", ql.Get(0))
	}
}

func TestQuickList_RemoveByVal_CountExceeds(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(1)
	ql.Add(2)
	removed := ql.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 100) // count > actual
	if removed != 2 {
		t.Errorf("应删除 2 个（不超过实际），实际 %d", removed)
	}
	if ql.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", ql.Len())
	}
}

func TestQuickList_RemoveByVal_Zero(t *testing.T) {
	// Redis LREM 语义：count=0 表示删除所有匹配元素
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(1)
	ql.Add(1)
	removed := ql.RemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 0)
	if removed != 3 {
		t.Errorf("count=0（Redis 语义：删除全部）应删除 3 个，实际 %d", removed)
	}
	if ql.Len() != 0 {
		t.Errorf("count=0 删除全部后 Len 应为 0，实际 %d", ql.Len())
	}
}

// ============================================================
// ReverseRemoveByVal — 右到左删除最多 count 个
// ============================================================

func TestQuickList_ReverseRemoveByVal_Partial(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 1, 1, 2, 3} {
		ql.Add(v)
	}
	removed := ql.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 1（从右），实际 %d", removed)
	}
	if ql.Len() != 3 {
		t.Errorf("Len 应为 3，实际 %d", ql.Len())
	}
	// 从左起只剩第一个 1
	if ql.Get(0) != 1 || ql.Get(1) != 2 || ql.Get(2) != 3 {
		t.Errorf("反向删除后应为 [1, 2, 3]，实际 [%v, %v, %v]",
			ql.Get(0), ql.Get(1), ql.Get(2))
	}
}

func TestQuickList_ReverseRemoveByVal_All(t *testing.T) {
	ql := NewQuickList()
	ql.Add("x")
	ql.Add("y")
	ql.Add("x")
	removed := ql.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(string) == "x"
	}, 2)
	if removed != 2 {
		t.Errorf("应删除 2 个 x，实际 %d", removed)
	}
	if ql.Len() != 1 {
		t.Errorf("Len 应为 1，实际 %d", ql.Len())
	}
}

func TestQuickList_ReverseRemoveByVal_CountExceeds(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(1)
	removed := ql.ReverseRemoveByVal(func(a interface{}) bool {
		return a.(int) == 1
	}, 5)
	if removed != 2 {
		t.Errorf("count > 实际时删除全部，实际 %d", removed)
	}
	if ql.Len() != 0 {
		t.Errorf("Len 应为 0，实际 %d", ql.Len())
	}
}

// ============================================================
// Len — 长度
// ============================================================

func TestQuickList_Len_AfterOperations(t *testing.T) {
	ql := NewQuickList()
	if ql.Len() != 0 {
		t.Errorf("初始 Len 应为 0，实际 %d", ql.Len())
	}
	ql.Add(1)
	ql.Add(2)
	if ql.Len() != 2 {
		t.Errorf("Add 后 Len 应为 2，实际 %d", ql.Len())
	}
	ql.Insert(1, "mid")
	if ql.Len() != 3 {
		t.Errorf("Insert 后 Len 应为 3，实际 %d", ql.Len())
	}
	ql.Remove(1)
	if ql.Len() != 2 {
		t.Errorf("Remove 后 Len 应为 2，实际 %d", ql.Len())
	}
	ql.RemoveLast()
	if ql.Len() != 1 {
		t.Errorf("RemoveLast 后 Len 应为 1，实际 %d", ql.Len())
	}
}

// ============================================================
// ForEach — 遍历
// ============================================================

func TestQuickList_ForEach_FullTraversal(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{10, 20, 30, 40} {
		ql.Add(v)
	}
	sum := 0
	count := 0
	ql.ForEach(func(i int, v interface{}) bool {
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

func TestQuickList_ForEach_EarlyBreak(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 2, 3, 4, 5} {
		ql.Add(v)
	}
	count := 0
	ql.ForEach(func(i int, v interface{}) bool {
		count++
		return v.(int) < 3
	})
	if count != 3 {
		t.Errorf("ForEach 应在 v=3 时中断，遍历次数 3，实际 %d", count)
	}
}

func TestQuickList_ForEach_Index(t *testing.T) {
	ql := NewQuickList()
	ql.Add("a")
	ql.Add("b")
	ql.Add("c")
	indices := make([]int, 0)
	ql.ForEach(func(i int, v interface{}) bool {
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

func TestQuickList_Contains_Found(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 2, 3, 4, 5} {
		ql.Add(v)
	}
	if !ql.Contains(func(a interface{}) bool {
		return a.(int) == 3
	}) {
		t.Error("Contains(3) 应为 true")
	}
}

func TestQuickList_Contains_NotFound(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	if ql.Contains(func(a interface{}) bool {
		return a.(int) == 999
	}) {
		t.Error("Contains(999) 应为 false")
	}
}

func TestQuickList_Contains_First(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	if !ql.Contains(func(a interface{}) bool {
		return a.(int) == 1
	}) {
		t.Error("Contains(1) 第一个元素 应为 true")
	}
}

func TestQuickList_Contains_Last(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	if !ql.Contains(func(a interface{}) bool {
		return a.(int) == 3
	}) {
		t.Error("Contains(3) 最后一个元素 应为 true")
	}
}

func TestQuickList_Contains_Empty(t *testing.T) {
	ql := NewQuickList()
	if ql.Contains(func(a interface{}) bool {
		return true
	}) {
		t.Error("空列表 Contains 应为 false")
	}
}

// ============================================================
// Range — 切片 [start, stop)
// ============================================================

func TestQuickList_Range_Normal(t *testing.T) {
	ql := NewQuickList()
	for _, v := range []int{1, 2, 3, 4, 5} {
		ql.Add(v)
	}
	slice := ql.Range(1, 4)
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

func TestQuickList_Range_Full(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	slice := ql.Range(0, ql.Len())
	if len(slice) != 3 {
		t.Errorf("Range(0,size) 长度应为 3，实际 %d", len(slice))
	}
}

func TestQuickList_Range_Single(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	slice := ql.Range(1, 2)
	if len(slice) != 1 || slice[0] != 2 {
		t.Errorf("Range(1,2) = %v, want [2]", slice)
	}
}

func TestQuickList_Range_StartEqualsStop(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	slice := ql.Range(1, 1)
	if len(slice) != 0 {
		t.Errorf("Range(1,1) 应为空切片，实际 %v", slice)
	}
}

func TestQuickList_Range_PanicOnNegativeStart(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(-1, 1) 应 panic")
		}
	}()
	ql.Range(-1, 1)
}

func TestQuickList_Range_PanicOnStartOutOfRange(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(2, 2) start>=size 应 panic")
		}
	}()
	ql.Range(2, 2)
}

func TestQuickList_Range_PanicOnStopBeforeStart(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	ql.Add(3)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(2, 1) 应 panic")
		}
	}()
	ql.Range(2, 1)
}

func TestQuickList_Range_PanicOnStopOutOfRange(t *testing.T) {
	ql := NewQuickList()
	ql.Add(1)
	ql.Add(2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Range(0, 3) 应 panic")
		}
	}()
	ql.Range(0, 3)
}

// ============================================================
// 跨页测试 — 数据量超过 pageSize (1024) 时的正确性
// ============================================================

func TestQuickList_Add_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 3000 // 跨越 2+ 个 page
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	if ql.Len() != n {
		t.Errorf("Len 应为 %d，实际 %d", n, ql.Len())
	}
	// 验证首尾和中间随机位置
	if ql.Get(0) != 0 {
		t.Errorf("Get(0) = %v, want 0", ql.Get(0))
	}
	if ql.Get(n-1) != n-1 {
		t.Errorf("Get(%d) = %v, want %d", n-1, ql.Get(n-1), n-1)
	}
	if ql.Get(1024) != 1024 {
		t.Errorf("Get(1024) = %v, want 1024（跨越页边界）", ql.Get(1024))
	}
	if ql.Get(2048) != 2048 {
		t.Errorf("Get(2048) = %v, want 2048", ql.Get(2048))
	}
}

func TestQuickList_RemoveLast_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	// 连续 RemoveLast 直到空
	for i := n - 1; i >= 0; i-- {
		val := ql.RemoveLast()
		if val != i {
			t.Fatalf("RemoveLast #%d = %v, want %d", n-i, val, i)
		}
	}
	if ql.Len() != 0 {
		t.Errorf("全部 RemoveLast 后 Len 应为 0，实际 %d", ql.Len())
	}
}

func TestQuickList_Remove_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	// 从头部删除
	val := ql.Remove(0)
	if val != 0 {
		t.Errorf("Remove(0) = %v, want 0", val)
	}
	if ql.Len() != n-1 {
		t.Errorf("Len 应为 %d，实际 %d", n-1, ql.Len())
	}
	if ql.Get(0) != 1 {
		t.Errorf("删除头后 Get(0) = %v, want 1", ql.Get(0))
	}
	// 从中间删除（第二页）
	val = ql.Remove(1023) // 原来的 1024，现在在 index 1023
	if val != 1024 {
		t.Errorf("Remove(1023) = %v, want 1024", val)
	}
	// 从尾部删除
	val = ql.Remove(ql.Len() - 1)
	if val != n-1 {
		t.Errorf("Remove(last) = %v, want %d", val, n-1)
	}
}

func TestQuickList_RemoveByVal_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	// 删除所有偶数
	removed := ql.RemoveByVal(func(a interface{}) bool {
		return a.(int)%2 == 0
	}, n) // count 足够大，全部删除
	if removed != 1000 {
		t.Errorf("应删除 1000 个偶数，实际 %d", removed)
	}
	if ql.Len() != 1000 {
		t.Errorf("Len 应为 1000，实际 %d", ql.Len())
	}
	// 剩余元素应全为奇数
	for i := 0; i < ql.Len(); i++ {
		if ql.Get(i).(int)%2 != 1 {
			t.Errorf("Get(%d) = %v, 应为奇数", i, ql.Get(i))
		}
	}
}

func TestQuickList_ForEach_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	count := 0
	sum := 0
	ql.ForEach(func(i int, v interface{}) bool {
		count++
		sum += v.(int)
		return true
	})
	if count != n {
		t.Errorf("ForEach 应遍历 %d 个元素，实际 %d", n, count)
	}
	expectedSum := (n - 1) * n / 2
	if sum != expectedSum {
		t.Errorf("sum 应为 %d，实际 %d", expectedSum, sum)
	}
}

func TestQuickList_Range_PageOverflow(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	// 跨页 Range
	slice := ql.Range(1000, 1050)
	if len(slice) != 50 {
		t.Errorf("Range(1000,1050) 长度应为 50，实际 %d", len(slice))
	}
	for i, v := range slice {
		if v != 1000+i {
			t.Errorf("Range[%d] = %v, want %d", i, v, 1000+i)
		}
	}
}

// ============================================================
// find — 内部二分查找优化（从两端遍历）
// ============================================================

func TestQuickList_Find_PageOptimization(t *testing.T) {
	ql := NewQuickList()
	n := 2000
	for i := 0; i < n; i++ {
		ql.Add(i)
	}
	// 前半部分：从 front 走
	if ql.Get(500) != 500 {
		t.Errorf("Get(500) = %v, want 500", ql.Get(500))
	}
	// 后半部分：从 back 走
	if ql.Get(1500) != 1500 {
		t.Errorf("Get(1500) = %v, want 1500", ql.Get(1500))
	}
}

// ============================================================
// 综合场景
// ============================================================

func TestQuickList_Comprehensive(t *testing.T) {
	ql := NewQuickList()

	// 构建列表: [10, 20, 30, 40, 50]
	ql.Add(10)
	ql.Add(20)
	ql.Add(30)
	ql.Add(40)
	ql.Add(50)

	if ql.Len() != 5 {
		t.Fatalf("初始 Len 应为 5，实际 %d", ql.Len())
	}

	// Insert: [10, 15, 20, 30, 40, 50]
	ql.Insert(1, 15)
	if ql.Get(1) != 15 {
		t.Errorf("Insert 后 Get(1) = %v, want 15", ql.Get(1))
	}

	// Set: [10, 15, 20, 30, 40, 60]
	ql.Set(ql.Len()-1, 60)
	if ql.Get(ql.Len()-1) != 60 {
		t.Errorf("Set 后尾部 = %v, want 60", ql.Get(ql.Len()-1))
	}

	// Remove: [10, 15, 30, 40, 60]
	val := ql.Remove(2)
	if val != 20 {
		t.Errorf("Remove(2) = %v, want 20", val)
	}

	// RemoveLast: [10, 15, 30, 40]
	val = ql.RemoveLast()
	if val != 60 {
		t.Errorf("RemoveLast = %v, want 60", val)
	}

	// RemoveAllByVal: 删除所有 40 → [10, 15, 30]
	removed := ql.RemoveAllByVal(func(a interface{}) bool {
		return a.(int) == 40
	})
	if removed != 1 {
		t.Errorf("RemoveAllByVal = %d, want 1", removed)
	}

	// 最终验证
	expected := []int{10, 15, 30}
	if ql.Len() != len(expected) {
		t.Fatalf("最终 Len 应为 %d，实际 %d", len(expected), ql.Len())
	}
	for i, want := range expected {
		if ql.Get(i) != want {
			t.Errorf("最终 Get(%d) = %v, want %d", i, ql.Get(i), want)
		}
	}
}
