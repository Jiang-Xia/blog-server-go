// Package captchasvg 封装图形验证码 SVG 生成，参数与 Nest svg-captcha 对齐并优化可读性。
package captchasvg

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	svgcaptcha "github.com/reu98/go-svg-captcha"
)

const (
	captchaWidth  uint16 = 100
	captchaHeight uint16 = 48
	fontSize      uint8  = 18
	curveCount    uint8  = 3
	noiseDots     int    = 66
)

var circleNoiseRe = regexp.MustCompile(`<circle[^>]*/>`)

// Create 生成 4 位图形验证码。
func Create() (*svgcaptcha.Result, error) {
	result, err := svgcaptcha.CreateByText(svgcaptcha.OptionText{
		Size:             4,
		Width:            captchaWidth,
		Height:           captchaHeight,
		FontSize:         fontSize,
		IsColor:          true,
		Curve:            curveCount,
		IgnoreCharacters: "0o1iIlL",
		CharactersPreset: "ABCDEFGHJKLMNPQRSTUVWXYZ23456789",
	})
	if err != nil {
		return nil, err
	}
	result.Data = replaceNoiseCircles(result.Data, captchaWidth, captchaHeight, noiseDots)
	return result, nil
}

// replaceNoiseCircles 去掉库默认圆点并按固定数量重新绘制。
func replaceNoiseCircles(svg string, width, height uint16, count int) string {
	svg = circleNoiseRe.ReplaceAllString(svg, "")
	if count <= 0 {
		return svg
	}

	var noise strings.Builder
	for i := 0; i < count; i++ {
		x := rand.Intn(int(width))
		y := rand.Intn(int(height))
		color := rand.Intn(0xFFFFFF)
		noise.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="1" fill="#%06X" />`, x, y, color))
	}
	return strings.Replace(svg, "</svg>", noise.String()+"</svg>", 1)
}
