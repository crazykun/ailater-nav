package utils

import (
	"crypto/md5"
	"fmt"
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
