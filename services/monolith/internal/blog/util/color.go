// Package util blog 域小工具。
package util

import (
	"fmt"
	"math/rand"
)

// RandomColor 生成分类/标签随机色，对齐 Nest getRandomClor。
func RandomColor() string {
	r := rand.Intn(256)
	g := rand.Intn(256)
	b := rand.Intn(256)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
