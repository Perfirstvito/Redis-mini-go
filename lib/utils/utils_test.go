package utils

import (
	"bytes"
	"testing"
)

func TestToCmdLine(t *testing.T) {
	got := ToCmdLine("SET", "key", "value")
	want := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	if !equalByteSlices(got, want) {
		t.Fatalf("ToCmdLine() = %q, want %q", got, want)
	}
}

func TestToCmdLineEmpty(t *testing.T) {
	got := ToCmdLine()
	if len(got) != 0 {
		t.Fatalf("ToCmdLine() = %q, want empty", got)
	}
}

func TestToCmdLine2(t *testing.T) {
	got := ToCmdLine2("GET", "a", "b")
	want := [][]byte{[]byte("GET"), []byte("a"), []byte("b")}
	if !equalByteSlices(got, want) {
		t.Fatalf("ToCmdLine2() = %q, want %q", got, want)
	}
}

func TestToCmdLine3(t *testing.T) {
	arg := []byte("raw")
	got := ToCmdLine3("SET", arg)
	want := [][]byte{[]byte("SET"), []byte("raw")}
	if !equalByteSlices(got, want) {
		t.Fatalf("ToCmdLine3() = %q, want %q", got, want)
	}
	// ToCmdLine3 must not copy: the argument slice should be the same backing array
	if &got[1][0] != &arg[0] {
		t.Fatal("ToCmdLine3 should keep the original []byte backing array")
	}
}

func TestBytesEquals(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty non-nil", []byte{}, []byte{}, true},
		{"nil vs empty", nil, []byte{}, false},
		{"empty vs nil", []byte{}, nil, false},
		{"equal content", []byte("abc"), []byte("abc"), true},
		{"different length", []byte("abc"), []byte("ab"), false},
		{"different content", []byte("abc"), []byte("abd"), false},
		{"both same pointer content", []byte("x"), []byte("x"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesEquals(tt.a, tt.b); got != tt.want {
				t.Fatalf("BytesEquals(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEqualsDispatchesByType(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"byte slices equal", []byte("abc"), []byte("abc"), true},
		{"byte slices differ", []byte("abc"), []byte("abd"), false},
		{"strings equal", "abc", "abc", true},
		{"strings differ", "abc", "abd", false},
		{"mixed types", []byte("abc"), "abc", false},
		{"ints equal", 1, 1, true},
		{"ints differ", 1, 2, false},
		{"both nil", nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equals(tt.a, tt.b); got != tt.want {
				t.Fatalf("Equals(%#v, %#v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestConvertRange(t *testing.T) {
	const size = int64(5) // valid indices 0..4
	tests := []struct {
		name       string
		start, end int64
		wantStart  int
		wantEnd    int
	}{
		{"full range", 0, 4, 0, 5},
		{"first element", 0, 0, 0, 1},
		{"single middle", 2, 2, 2, 3},
		{"negative end is last", 0, -1, 0, 5},
		{"negative start and end", -2, -1, 3, 5},
		{"end clamped to size", 0, 10, 0, 5},
		{"start beyond size", 5, 6, -1, -1},
		{"start too negative", -6, -1, -1, -1},
		{"end too negative", 0, -6, -1, -1},
		{"end before start", 3, 1, -1, -1},
		{"negative end before start", -1, 0, -1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := ConvertRange(tt.start, tt.end, size)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Fatalf("ConvertRange(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.start, tt.end, size, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestConvertRangeZeroSize(t *testing.T) {
	// any positive index is out of bound when size == 0
	if start, end := ConvertRange(0, 0, 0); start != -1 || end != -1 {
		t.Fatalf("ConvertRange(0,0,0) = (%d,%d), want (-1,-1)", start, end)
	}
}

func TestRemoveDuplicatesPreservesOrder(t *testing.T) {
	in := [][]byte{
		[]byte("a"), []byte("b"), []byte("a"), []byte("c"), []byte("b"), []byte("d"),
	}
	got := RemoveDuplicates(in)
	want := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	if !equalByteSlices(got, want) {
		t.Fatalf("RemoveDuplicates() = %q, want %q", got, want)
	}
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	got := RemoveDuplicates(nil)
	if len(got) != 0 {
		t.Fatalf("RemoveDuplicates(nil) = %q, want empty", got)
	}
}

func TestRandStringLengthAndCharset(t *testing.T) {
	s := RandString(32)
	if len(s) != 32 {
		t.Fatalf("RandString(32) len = %d, want 32", len(s))
	}
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i, c := range s {
		if !bytes.ContainsRune([]byte(allowed), c) {
			t.Fatalf("RandString() char %q at %d not in allowed charset", c, i)
		}
	}
	// different calls are very likely to differ
	if RandString(40) == RandString(40) {
		t.Fatal("RandString(40) returned identical values twice")
	}
	if RandString(0) != "" {
		t.Fatalf("RandString(0) = %q, want empty", RandString(0))
	}
}

func TestRandHexStringCharset(t *testing.T) {
	s := RandHexString(24)
	if len(s) != 24 {
		t.Fatalf("RandHexString(24) len = %d, want 24", len(s))
	}
	const allowed = "0123456789abcdef"
	for i, c := range s {
		if !bytes.ContainsRune([]byte(allowed), c) {
			t.Fatalf("RandHexString() char %q at %d not hex", c, i)
		}
	}
}

func TestRandIndexIsPermutation(t *testing.T) {
	const size = 10
	idx := RandIndex(size)
	if len(idx) != size {
		t.Fatalf("RandIndex(10) len = %d, want 10", len(idx))
	}
	seen := make(map[int]bool)
	for _, v := range idx {
		if v < 0 || v >= size {
			t.Fatalf("RandIndex returned out-of-range value %d", v)
		}
		if seen[v] {
			t.Fatalf("RandIndex returned duplicate %d", v)
		}
		seen[v] = true
	}
}

func equalByteSlices(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
