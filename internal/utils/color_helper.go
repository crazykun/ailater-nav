package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// GenerateColorFromName 根据网站名称生成一致的颜色代码
func GenerateColorFromName(name string) string {
	// 使用MD5生成一个哈希值
	hash := md5.Sum([]byte(name))

	// 将哈希的前几个字节转换为整数
	// 使用哈希值来确定色调(Hue)，确保每个名字都有固定的颜色
	hue := int(hash[0])*256 + int(hash[1])

	// 计算色调角度 (0-360度)
	h := float64(hue % 360)

	// 使用固定的饱和度和亮度来确保颜色看起来不错
	s := 70.0 // 饱和度 70%
	l := 65.0 // 亮度 65%

	// 将HSL转换为RGB
	r, g, b := hslToRgb(h, s, l)

	// 转换为十六进制颜色代码
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hslToRgb 将HSL颜色值转换为RGB
func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	h = h / 360.0
	s = s / 100.0
	l = l / 100.0

	var r, g, b float64
	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRgb(p, q, h+1.0/3)
		g = hueToRgb(p, q, h)
		b = hueToRgb(p, q, h-1.0/3)
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

// hueToRgb 辅助函数，用于HSL到RGB转换
func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2 {
		return q
	}
	if t < 2.0/3 {
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}

// GetInitialsFromName 从网站名称获取首字母，用于显示
func GetInitialsFromName(name string) string {
	if len(name) == 0 {
		return "?"
	}

	// 获取第一个字符作为首字母
	// 如果是多字节字符（如中文），则获取第一个字符
	runes := []rune(name)
	if len(runes) == 0 {
		return "?"
	}

	// 检查是否有多个英文单词（包含空格）
	if len(name) > 3 {
		// 查找第一个空格
		spaceIndex := -1
		for i, char := range name {
			if char == ' ' {
				spaceIndex = i
				break
			}
		}

		// 如果有空格，取前两个单词的首字母
		if spaceIndex > 0 && spaceIndex < len(name)-1 {
			firstChar := string(name[0])
			secondChar := string(name[spaceIndex+1])
			return firstChar + secondChar
		}
	}

	// 取第一个字符
	initial := string(runes[0])
	// 如果名称较长，添加第二个字符作为补充
	if len(runes) > 3 && len(runes) >= 2 {
		initial = string(runes[0:2])
	}

	return initial
}

var tagPalette = []string{
	"bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-300 dark:border-emerald-800",
	"bg-sky-50 text-sky-700 border border-sky-200 dark:bg-sky-900/20 dark:text-sky-300 dark:border-sky-800",
	"bg-violet-50 text-violet-700 border border-violet-200 dark:bg-violet-900/20 dark:text-violet-300 dark:border-violet-800",
	"bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:border-amber-800",
	"bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-900/20 dark:text-rose-300 dark:border-rose-800",
	"bg-teal-50 text-teal-700 border border-teal-200 dark:bg-teal-900/20 dark:text-teal-300 dark:border-teal-800",
	"bg-indigo-50 text-indigo-700 border border-indigo-200 dark:bg-indigo-900/20 dark:text-indigo-300 dark:border-indigo-800",
	"bg-orange-50 text-orange-700 border border-orange-200 dark:bg-orange-900/20 dark:text-orange-300 dark:border-orange-800",
}

func GetTagColorClass(tag string) string {
	normalized := strings.TrimSpace(tag)
	if normalized == "" {
		return tagPalette[0]
	}

	hash := md5.Sum([]byte(normalized))
	index := int(hash[0]) % len(tagPalette)
	return tagPalette[index]
}
