package timewheel

import (
	"testing"
	"time"
)

// ============================================================
// Delay
// ============================================================

func TestDelay(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		key      string
	}{
		{name: "正常延迟", duration: 5 * time.Second, key: "test-delay"},
		{name: "零延迟", duration: 0, key: "test-zero"},
		{name: "毫秒级延迟", duration: 100 * time.Millisecond, key: "test-ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 不应 panic
			Delay(tt.duration, tt.key, func() {})
		})
	}
}

func TestDelayNegative(t *testing.T) {
	// 负延迟：AddJob 内部跳过，不应 panic
	Delay(-time.Second, "test-negative", func() {})
	Delay(-time.Millisecond, "test-negative-ms", func() {})
}

func TestDelayEmptyKey(t *testing.T) {
	// 空 key 不应 panic
	Delay(time.Second, "", func() {})
}

// ============================================================
// At
// ============================================================

func TestAt(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
		key  string
	}{
		{name: "未来时间", at: time.Now().Add(2 * time.Second), key: "test-at-future"},
		{name: "未来毫秒", at: time.Now().Add(50 * time.Millisecond), key: "test-at-ms"},
		{name: "远未来时间", at: time.Now().Add(time.Hour), key: "test-at-hour"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 不应 panic
			At(tt.at, tt.key, func() {})
		})
	}
}

func TestAtPastTime(t *testing.T) {
	// 过去时间：sub 为负，AddJob 内部直接 return，不应 panic
	past := time.Now().Add(-time.Hour)
	At(past, "test-past", func() {})
}

func TestAtExactNow(t *testing.T) {
	// 正好是当前时间（sub 接近 0），不应 panic
	At(time.Now(), "test-now", func() {})
}

func TestAtEmptyKey(t *testing.T) {
	future := time.Now().Add(time.Second)
	At(future, "", func() {})
}

// ============================================================
// Cancel
// ============================================================

func TestCancel(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "正常取消", key: "cancel-me"},
		{name: "取消后再次取消同一key", key: "cancel-twice"},
		{name: "取消未添加的key", key: "never-added"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先添加再取消
			Delay(10*time.Second, tt.key, func() {})
			Cancel(tt.key)
			Cancel(tt.key) // 再次取消不应 panic
		})
	}
}

func TestCancelEmptyKey(t *testing.T) {
	// 空 key：RemoveJob 内部直接 return，不应 panic
	Cancel("")
}

func TestCancelNotFound(t *testing.T) {
	// 不存在的 key：RemoveJob 内部 timer 没有该 key，无操作，不应 panic
	Cancel("not-exist")
}

// ============================================================
// Delay + Cancel 组合
// ============================================================

func TestDelayAndCancel(t *testing.T) {
	Delay(30*time.Second, "combo-1", func() {})
	Cancel("combo-1")

	// 多次组合
	Delay(time.Minute, "combo-2", func() {})
	Delay(time.Minute, "combo-2", func() {}) // 重复 key 替换
	Cancel("combo-2")
}
