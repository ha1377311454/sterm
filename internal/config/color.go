package config

import "github.com/gdamore/tcell/v2"

// Color 封装颜色字符串：十六进制 "#rrggbb"、命名颜色或 "default"/"-"。
type Color string

// Color 将字符串颜色转换为 tcell.Color。
func (c Color) Color() tcell.Color {
	if c == "" || c == "default" || c == "-" {
		return tcell.ColorDefault
	}
	return tcell.GetColor(string(c))
}
