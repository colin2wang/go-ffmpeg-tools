package ffmpeg

import (
	"path/filepath"
	"strings"
)

// langRule 语言推断规则：tokens 为 ASCII 缩写（整词匹配，两侧须为非字母数字，
// 避免英文文件名中的 "desc"/"scene" 之类误判为 chs）；substrs 为 CJK 等无需边界的子串。
// 规则按序检查，靠前优先。
var langRules = []struct {
	lang    string
	tokens  []string
	substrs []string
}{
	{"chs", []string{"chs", "zh-hans", "zh-cn", "sc", "gb", "gbk", "simplified"}, []string{"简体", "简中"}},
	{"cht", []string{"cht", "zh-hant", "zh-tw", "zh-hk", "tc", "big5", "traditional"}, []string{"繁體", "繁体", "繁中"}},
	{"jpn", []string{"jpn", "jp", "japanese"}, []string{"日文", "日语", "日本語", "日本语"}},
	{"kor", []string{"kor", "kr", "korean"}, []string{"韩文", "韩语", "韓文", "한국"}},
	{"eng", []string{"eng", "english"}, []string{"英文", "英语"}},
}

// InferLang 从字幕文件名推断语言代码（chs/cht/eng/jpn/kor）；无匹配返回默认值 eng。
func InferLang(filename string) string {
	base := filepath.Base(filename)
	name := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))
	for _, rule := range langRules {
		for _, tok := range rule.tokens {
			if matchWord(name, tok) {
				return rule.lang
			}
		}
		for _, s := range rule.substrs {
			if strings.Contains(name, s) {
				return rule.lang
			}
		}
	}
	return "eng"
}

// matchWord 整词匹配：tok 每次出现位置的两侧均为非 ASCII 小写字母/数字（或字符串边界）
func matchWord(s, tok string) bool {
	for i := 0; ; {
		idx := strings.Index(s[i:], tok)
		if idx < 0 {
			return false
		}
		start := i + idx
		end := start + len(tok)
		if langBoundary(s, start-1) && langBoundary(s, end) {
			return true
		}
		i = start + 1
	}
}

// langBoundary 判断 s[i] 是否为词边界：越界（字符串首尾）或非 [a-z0-9]
func langBoundary(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return true
	}
	c := s[i]
	return !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9')
}
