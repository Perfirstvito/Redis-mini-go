package wildcard

import (
	"strings"
	"testing"
)

// TestCompilePattern 测试通配符编译（正常 + 错误路径）
func TestCompilePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
		errMsg  string // 期望的错误信息片段，仅 wantErr=true 时检查
	}{
		// ========== 正常编译 ==========
		{name: "单星号 *", pattern: "*"},
		{name: "单问号 ?", pattern: "?"},
		{name: "普通字符串", pattern: "hello"},
		{name: "Redis风格 h?llo", pattern: "h?llo"},
		{name: "Redis风格 h*llo", pattern: "h*llo"},
		{name: "字符集 [abc]", pattern: "h[ae]llo"},
		{name: "排除字符集 [^abc]", pattern: "[^abc]"},
		{name: "范围字符集 [a-z]", pattern: "[a-z]"},
		{name: "转义星号 \\*", pattern: `\*`},
		{name: "转义问号 \\?", pattern: `\?`},
		{name: "转义反斜杠 \\\\", pattern: `\\`},
		{name: "转义普通字符 \\a", pattern: `\a`},
		{name: "开头 ^ 作为字面量", pattern: "^hello"},
		{name: "^ 在 [] 内作为排除", pattern: "h[^ae]llo"},
		{name: "\\[^ 中的 ^ 作为字面量", pattern: `\[^abc`},
		{name: "复杂组合 * ? [a-z]", pattern: "file-*-v?[0-9].txt"},
		{name: "regex特殊字符作为字面量: +", pattern: "+"},
		{name: "regex特殊字符作为字面量: )", pattern: ")"},
		{name: "regex特殊字符作为字面量: $", pattern: "$"},
		{name: "regex特殊字符作为字面量: .", pattern: "."},
		{name: "regex特殊字符作为字面量: {", pattern: "{"},
		{name: "regex特殊字符作为字面量: }", pattern: "}"},
		{name: "regex特殊字符作为字面量: |", pattern: "|"},

		// ========== 错误 ==========
		{name: "末尾反斜杠报错", pattern: `abc\`, wantErr: true, errMsg: "end with escape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompilePattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CompilePattern(%q) expected error, got nil", tt.pattern)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("CompilePattern(%q) error = %q, want containing %q",
						tt.pattern, err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("CompilePattern(%q) unexpected error: %v", tt.pattern, err)
				return
			}
			if got == nil {
				t.Errorf("CompilePattern(%q) returned nil Pattern without error", tt.pattern)
			}
		})
	}
}

// TestIsMatch 测试通配符匹配逻辑
func TestIsMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		// ===================== * 通配符 =====================
		{name: "* 匹配任意字符串", pattern: "*", input: "anything", want: true},
		{name: "* 匹配空字符串", pattern: "*", input: "", want: true},
		{name: "h*llo 匹配 hello", pattern: "h*llo", input: "hello", want: true},
		{name: "h*llo 匹配 hllo", pattern: "h*llo", input: "hllo", want: true},
		{name: "h*llo 匹配 heeeello", pattern: "h*llo", input: "heeeello", want: true},
		{name: "h*llo 不匹配 halo", pattern: "h*llo", input: "halo", want: false},
		{name: "*llo 匹配 hello", pattern: "*llo", input: "hello", want: true},
		{name: "he* 匹配 hello", pattern: "he*", input: "hello", want: true},
		{name: "he* 匹配 he", pattern: "he*", input: "he", want: true},
		{name: "he* 不匹配 ha", pattern: "he*", input: "ha", want: false},
		{name: "*a*b* 匹配 axbxc", pattern: "*a*b*", input: "axbxc", want: true},

		// ===================== ? 通配符 =====================
		{name: "? 匹配单个字符", pattern: "?", input: "a", want: true},
		{name: "? 不匹配空字符串", pattern: "?", input: "", want: false},
		{name: "? 不匹配两个字符", pattern: "?", input: "ab", want: false},
		{name: "h?llo 匹配 hello", pattern: "h?llo", input: "hello", want: true},
		{name: "h?llo 匹配 hallo", pattern: "h?llo", input: "hallo", want: true},
		{name: "h?llo 不匹配 hllo（缺一个字符）", pattern: "h?llo", input: "hllo", want: false},
		{name: "h?llo 不匹配 heello（多一个字符）", pattern: "h?llo", input: "heello", want: false},
		{name: "?.txt 匹配 a.txt", pattern: "?.txt", input: "a.txt", want: true},
		{name: "?.txt 不匹配 ab.txt", pattern: "?.txt", input: "ab.txt", want: false},
		{name: "?.txt 不匹配 .txt", pattern: "?.txt", input: ".txt", want: false},

		// ===================== 字符集 [...] =====================
		{name: "[abc] 匹配 a", pattern: "[abc]", input: "a", want: true},
		{name: "[abc] 匹配 b", pattern: "[abc]", input: "b", want: true},
		{name: "[abc] 匹配 c", pattern: "[abc]", input: "c", want: true},
		{name: "[abc] 不匹配 d", pattern: "[abc]", input: "d", want: false},
		{name: "[abc] 不匹配空字符串", pattern: "[abc]", input: "", want: false},
		{name: "[abc] 不匹配 ab", pattern: "[abc]", input: "ab", want: false},
		{name: "h[ae]llo 匹配 hello", pattern: "h[ae]llo", input: "hello", want: true},
		{name: "h[ae]llo 匹配 hallo", pattern: "h[ae]llo", input: "hallo", want: true},
		{name: "h[ae]llo 不匹配 hillo", pattern: "h[ae]llo", input: "hillo", want: false},
		{name: "h[ae]llo 不匹配 hllo", pattern: "h[ae]llo", input: "hllo", want: false},

		// ===================== 排除字符集 [^...] =====================
		{name: "[^abc] 匹配 d", pattern: "[^abc]", input: "d", want: true},
		{name: "[^abc] 匹配数字", pattern: "[^abc]", input: "1", want: true},
		{name: "[^abc] 不匹配 a", pattern: "[^abc]", input: "a", want: false},
		{name: "[^abc] 不匹配 b", pattern: "[^abc]", input: "b", want: false},
		{name: "[^abc] 不匹配空字符串", pattern: "[^abc]", input: "", want: false},
		{name: "h[^e]llo 匹配 hallo", pattern: "h[^e]llo", input: "hallo", want: true},
		{name: "h[^e]llo 匹配 hillo", pattern: "h[^e]llo", input: "hillo", want: true},
		{name: "h[^e]llo 不匹配 hello", pattern: "h[^e]llo", input: "hello", want: false},

		// ===================== 范围字符集 [a-z] =====================
		{name: "[a-z] 匹配小写字母", pattern: "[a-z]", input: "m", want: true},
		{name: "[a-z] 不匹配数字", pattern: "[a-z]", input: "5", want: false},
		{name: "[a-z] 不匹配大写", pattern: "[a-z]", input: "M", want: false},
		{name: "[0-9] 匹配数字", pattern: "[0-9]", input: "7", want: true},
		{name: "[0-9] 不匹配字母", pattern: "[0-9]", input: "a", want: false},
		{name: "多个范围 [a-zA-Z]", pattern: "[a-zA-Z]", input: "X", want: true},
		{name: "多个范围 [a-zA-Z] 匹配小写", pattern: "[a-zA-Z]", input: "x", want: true},
		{name: "多个范围 [a-zA-Z] 不匹配数字", pattern: "[a-zA-Z]", input: "3", want: false},

		// ===================== 转义序列 =====================
		{name: "\\* 匹配字面量 *", pattern: `\*`, input: "*", want: true},
		{name: "\\* 不匹配 a", pattern: `\*`, input: "a", want: false},
		{name: "\\* 不匹配多个字符", pattern: `\*`, input: "abc", want: false},
		{name: "\\? 匹配字面量 ?", pattern: `\?`, input: "?", want: true},
		{name: "\\? 不匹配 a", pattern: `\?`, input: "a", want: false},
		{name: "\\\\ 匹配字面量 \\", pattern: `\\`, input: `\`, want: true},
		{name: "\\\\ 不匹配 a", pattern: `\\`, input: "a", want: false},
		{name: "转义混合 \\*-\\? 匹配 *-?", pattern: `\*-\?`, input: "*-?", want: true},

		// ===================== Regex 特殊字符作为字面量 =====================
		{name: "+ 作为字面量", pattern: "+", input: "+", want: true},
		{name: "+ 不匹配 a", pattern: "+", input: "a", want: false},
		{name: ") 作为字面量", pattern: ")", input: ")", want: true},
		{name: "$ 作为字面量", pattern: "$", input: "$", want: true},
		{name: ". 作为字面量（不匹配任意字符）", pattern: ".", input: "a", want: false},
		{name: ". 匹配字面量点", pattern: ".", input: ".", want: true},
		{name: "{ 作为字面量", pattern: "{", input: "{", want: true},
		{name: "} 作为字面量", pattern: "}", input: "}", want: true},
		{name: "| 作为字面量", pattern: "|", input: "|", want: true},

		// ===================== ^ 处理 =====================
		{name: "^hello 匹配 ^hello", pattern: "^hello", input: "^hello", want: true},
		{name: "^hello 不匹配 hello（^ 是字面量）", pattern: "^hello", input: "hello", want: false},
		{name: "a^b 匹配 a^b（中间 ^ 是字面量）", pattern: "a^b", input: "a^b", want: true},
		{name: "a^b 不匹配 ab", pattern: "a^b", input: "ab", want: false},

		// ===================== 锚定 =====================
		{name: "精确匹配 hello", pattern: "hello", input: "hello", want: true},
		{name: "部分匹配 xhellox 失败", pattern: "hello", input: "xhellox", want: false},
		{name: "前缀匹配 he* 成功", pattern: "he*", input: "hello", want: true},
		{name: "后缀匹配 *llo 成功", pattern: "*llo", input: "hello", want: true},
		{name: "空模式匹配空串", pattern: "", input: "", want: true},
		{name: "空模式不匹配非空", pattern: "", input: "a", want: false},

		// ===================== 边界 =====================
		{name: "多个 * 连续", pattern: "a***b", input: "ab", want: true},
		{name: "多个 * 连续匹配中段", pattern: "a***b", input: "aXYZb", want: true},
		{name: "多个 ? 连续", pattern: "???", input: "abc", want: true},
		{name: "多个 ? 不够", pattern: "???", input: "ab", want: false},
		{name: "*? 组合", pattern: "*?", input: "abc", want: true},
		{name: "*? 最少一个字符", pattern: "*?", input: "", want: false},
		{name: "纯中文匹配", pattern: "你好*世界", input: "你好美丽的世界", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := CompilePattern(tt.pattern)
			if err != nil {
				t.Fatalf("CompilePattern(%q) 编译失败: %v", tt.pattern, err)
			}
			got := p.IsMatch(tt.input)
			if got != tt.want {
				t.Errorf("IsMatch(%q, %q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}

// TestIsMatchNilPattern 测试 nil Pattern 的行为（防御性）
func TestIsMatchNilPattern(t *testing.T) {
	var p *Pattern
	// nil receiver 调用 IsMatch 会 panic，记录当前行为
	defer func() {
		if r := recover(); r != nil {
			t.Logf("nil Pattern.IsMatch 触发 panic（符合预期）: %v", r)
		}
	}()
	p.IsMatch("hello")
}
