package tui

import (
	"fmt"
	"strings"
)

var BannerBlockTech = []string{
	"  ██████╗███╗   ██╗ ██████╗ ███████╗███████╗██╗     ",
	"  ██╔═══╝████╗  ██║██╔════╝ ██╔════╝██╔════╝██║     ",
	"  █████╗ ██╔██╗ ██║██║  ███╗███████╗█████╗  ██║     ",
	"  ██╔══╝ ██║╚██╗██║██║   ██║╚════██║██╔══╝  ██║     ",
	"  ██████╗██║ ╚████║╚██████╔╝███████║███████╗███████╗",
	"  ╚═════╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚══════╝╚══════╝",
}

var Highlights = map[string]bool{
	"1": true, "2": true, "6": true, "11": true, "S": true, "s": true,
}

func getW(w []int) int {
	if len(w) > 0 && w[0] > 0 {
		return w[0]
	}
	return Width
}

func PLine(contentStr string, plainLen int, width ...int) string {
	w := getW(width)
	pad := (w - 2) - plainLen
	if pad < 0 {
		pad = 0
	}
	return CBor("│") + contentStr + BgSpaces(pad) + CBor("│") + CEnd
}

func RenderBoxTop(title string, width ...int) string {
	w := getW(width)
	if title != "" {
		tClean := " " + title + " "
		tLen := len(tClean)
		rem := w - 2 - tLen - 2
		if rem < 0 {
			rem = 0
		}
		return Border("┌──") + Primary(tClean) + Border(strings.Repeat("─", rem)+"┐")
	}
	return Border("┌" + strings.Repeat("─", w-2) + "┐")
}

func RenderBoxBottom(footer string, width ...int) string {
	w := getW(width)
	if footer != "" {
		fClean := " " + footer + " "
		fLen := len(fClean)
		rem := w - 2 - fLen - 2
		if rem < 0 {
			rem = 0
		}
		return Border("└──") + TextDim(fClean) + Border(strings.Repeat("─", rem)+"┘")
	}
	return Border("└" + strings.Repeat("─", w-2) + "┘")
}

func RenderBoxDivider(title string, width ...int) string {
	w := getW(width)
	if title != "" {
		tClean := " " + title + " "
		tLen := len(tClean)
		rem := w - 2 - tLen - 2
		if rem < 0 {
			rem = 0
		}
		return Border("├──") + Primary(tClean) + Border(strings.Repeat("─", rem)+"┤")
	}
	return Border("├" + strings.Repeat("─", w-2) + "┤")
}

func RenderEmptyLine(width ...int) string {
	return PLine("", 0, width...)
}

func RenderHdrLine(title string, width ...int) string {
	content := BgSpaces(2) + CPri("▸ "+title)
	return PLine(content, 2+len("▸ "+title), width...)
}

func RenderSubHdrLine(title string, width ...int) string {
	content := BgSpaces(2) + CMut("• "+title)
	return PLine(content, 2+len("• "+title), width...)
}

func RenderHeader(title string, breadcrumb string, width ...int) {
	if breadcrumb == "" {
		breadcrumb = "engsel"
	}
	w := getW(width)
	fmt.Println(RenderBoxTop("", w))
	bcStr := fmt.Sprintf("  %s ❯ %s", breadcrumb, title)
	content := BgSpaces(2) + CDim(breadcrumb+" ") + CPri("❯ ") + CWhi(title)
	fmt.Println(PLine(content, len(bcStr), w))
	fmt.Println(RenderBoxDivider("", w))
}

func RenderKVLine(label, val string, valFn func(string) string, width ...int) string {
	w := getW(width)
	lblW := 12
	lblStr := fmt.Sprintf("  %-*s: ", lblW, label)
	valClean := TruncateStr(val, w-4-len(lblStr))
	content := BgSpaces(2) + CMut(fmt.Sprintf("%-*s: ", lblW, label)) + valFn(valClean)
	plainLen := len(lblStr) + CleanLen(valClean)
	return PLine(content, plainLen, w)
}

func RenderKV2Col(lLabel, lVal string, lFn func(string) string, rLabel, rVal string, rFn func(string) string, col2 int, width ...int) string {
	w := getW(width)
	lblW := 8
	leftPlain := fmt.Sprintf("  %-*s: %s", lblW, lLabel, lVal)
	pad := col2 - len(leftPlain)
	if pad < 1 {
		pad = 1
	}
	rightPlain := fmt.Sprintf("%-*s: %s", lblW, rLabel, rVal)

	content := BgSpaces(2) +
		CMut(fmt.Sprintf("%-*s: ", lblW, lLabel)) +
		lFn(lVal) +
		BgSpaces(pad) +
		CMut(fmt.Sprintf("%-*s: ", lblW, rLabel)) +
		rFn(rVal)
	plainLen := len(leftPlain) + pad + len(rightPlain)
	return PLine(content, plainLen, w)
}

func RenderCardRow(lLabel, lVal string, lFn func(string) string, rLabel, rVal string, rFn func(string) string, col2 int, width ...int) string {
	return RenderKV2Col(lLabel, lVal, lFn, rLabel, rVal, rFn, col2, width...)
}

func RenderMenuRow(k1, t1, k2, t2 string, col1W int, padKey bool, width ...int) string {
	w := getW(width)
	hi1 := Highlights[k1]
	hi2 := Highlights[k2]

	c1 := CMut(t1)
	if hi1 {
		c1 = CWhi(t1)
	}
	c2 := CMut(t2)
	if hi2 {
		c2 = CWhi(t2)
	}

	k1Fmt := k1
	k2Fmt := k2
	if padKey {
		k1Fmt = fmt.Sprintf("%2s", k1)
		k2Fmt = fmt.Sprintf("%2s", k2)
	}

	b1 := CBor("[") + CPri(k1Fmt) + CBor("]")
	b2 := CBor("[") + CPri(k2Fmt) + CBor("]")

	plainLeft := fmt.Sprintf("  [%s] %s", k1Fmt, t1)
	pad := col1W - len(plainLeft)
	if pad < 1 {
		pad = 1
	}

	content := BgSpaces(2) + b1 + BgSpaces(1) + c1 + BgSpaces(pad) + b2 + BgSpaces(1) + c2
	plainTotal := len(plainLeft) + pad + len(fmt.Sprintf("[%s] %s", k2Fmt, t2))
	return PLine(content, plainTotal, w)
}

func RenderAlert(msg string, level string, width ...int) {
	w := getW(width)
	var icon, colorTag string
	var colFn func(string) string

	switch strings.ToLower(level) {
	case "success":
		icon = "✔"
		colorTag = "SUKSES"
		colFn = CSuc
	case "error":
		icon = "✖"
		colorTag = "ERROR"
		colFn = CDan
	case "warning":
		icon = "⚠"
		colorTag = "PERINGATAN"
		colFn = CAmb
	default:
		icon = "ℹ"
		colorTag = "INFO"
		colFn = CPri
	}

	boxHeader := fmt.Sprintf("[%s %s]", icon, colorTag)
	maxContentW := w - 6 - len(boxHeader)
	msgTrunc := TruncateStr(msg, maxContentW)

	plain := fmt.Sprintf("  %s %s", boxHeader, msgTrunc)
	content := BgSpaces(2) + colFn(boxHeader) + BgSpaces(1) + CWhi(msgTrunc)

	fmt.Println()
	fmt.Println(RenderBoxTop(colorTag, w))
	fmt.Println(PLine(content, len(plain), w))
	fmt.Println(RenderBoxBottom("", w))
}

func RenderBanner(subtitle string, width ...int) {
	w := getW(width)
	for _, line := range BannerBlockTech {
		fmt.Println(Primary(line))
	}
	subCentered := subtitle + "  ·  v2.0"
	pad := (w - len(subCentered)) / 2
	if pad < 0 {
		pad = 0
	}
	fmt.Println(TextDim(strings.Repeat(" ", pad) + subCentered))
	fmt.Println()
}
