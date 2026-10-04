package tui

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Background Canvas (Dark Slate Interior: #181e28)
const (
	BGR = 24
	BGG = 30
	BGB = 40

	HBGR = 40
	HBGG = 50
	HBGB = 68

	DefaultWidth = 56
)

var Width = DefaultWidth

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func Style(fr, fg, fb int, text string, bold bool) string {
	b := ""
	if bold {
		b = "1;"
	}
	return fmt.Sprintf("\033[%s38;2;%d;%d;%d;48;2;%d;%d;%dm%s", b, fr, fg, fb, BGR, BGG, BGB, text)
}

func BgSpaces(count int) string {
	if count < 0 {
		count = 0
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm%s", BGR, BGG, BGB, strings.Repeat(" ", count))
}

func RGB(r, g, b int, text string) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}

func BoldRGB(r, g, b int, text string) string {
	return fmt.Sprintf("\033[1;38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}

// Semantic Palette (Standalone)
func Primary(t string) string   { return BoldRGB(56, 189, 248, t) }  // #38bdf8
func Secondary(t string) string { return BoldRGB(148, 163, 184, t) } // #94a3b8
func Success(t string) string   { return BoldRGB(52, 211, 153, t) }  // #34d399
func Amber(t string) string     { return BoldRGB(251, 191, 36, t) }  // #fbbf24
func Danger(t string) string    { return BoldRGB(248, 113, 113, t) } // #f87171
func Purple(t string) string    { return BoldRGB(192, 132, 252, t) } // #c084fc
func TextWhite(t string) string { return BoldRGB(248, 250, 252, t) } // #f8fafc
func TextMuted(t string) string { return RGB(148, 163, 184, t) }      // #94a3b8
func TextDim(t string) string   { return RGB(100, 116, 139, t) }      // #64748b
func Border(t string) string    { return RGB(71, 85, 105, t) }        // #475569

// Interior Container Stylers (locked canvas background)
func CPri(t string) string { return Style(56, 189, 248, t, true) }
func CSec(t string) string { return Style(148, 163, 184, t, true) }
func CSuc(t string) string { return Style(52, 211, 153, t, true) }
func CAmb(t string) string { return Style(251, 191, 36, t, true) }
func CDan(t string) string { return Style(248, 113, 113, t, true) }
func CPur(t string) string { return Style(192, 132, 252, t, true) }
func CWhi(t string) string { return Style(248, 250, 252, t, true) }
func CMut(t string) string { return Style(148, 163, 184, t, false) }
func CDim(t string) string { return Style(100, 116, 139, t, false) }
func CBor(t string) string { return Style(71, 85, 105, t, false) }

const CEnd = "\033[0m"

func CleanLen(s string) int {
	clean := ansiRegex.ReplaceAllString(s, "")
	w := 0
	for _, r := range clean {
		if r >= 0x1100 &&
			(r <= 0x115f || r == 0x2329 || r == 0x232a ||
				(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
				(r >= 0xac00 && r <= 0xd7a3) ||
				(r >= 0xf900 && r <= 0xfaff) ||
				(r >= 0xfe10 && r <= 0xfe19) ||
				(r >= 0xfe30 && r <= 0xfe6f) ||
				(r >= 0xff00 && r <= 0xff60) ||
				(r >= 0xffe0 && r <= 0xffe6) ||
				(r >= 0x20000 && r <= 0x2fffd) ||
				(r >= 0x30000 && r <= 0x3fffd)) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func TruncateStr(text string, maxW int, ellipsis ...string) string {
	el := "…"
	if len(ellipsis) > 0 {
		el = ellipsis[0]
	}
	if CleanLen(text) <= maxW {
		return text
	}
	target := maxW - CleanLen(el)
	if target <= 0 {
		return el
	}

	curW := 0
	var res []rune
	for _, r := range text {
		rw := 1
		if utf8.RuneLen(r) > 1 && CleanLen(string(r)) == 2 {
			rw = 2
		}
		if curW+rw > target {
			break
		}
		curW += rw
		res = append(res, r)
	}
	return string(res) + el
}

func Badge(text string, colorFn func(string) string) string {
	return CBor("[") + colorFn(text) + CBor("]")
}

func StatusBadge(status string) string {
	s := strings.ToUpper(strings.TrimSpace(status))
	switch s {
	case "SUCCESS", "SUCCEEDED", "AKTIF", "ACTIVE", "VALID", "PAID":
		return CSuc(fmt.Sprintf("[%s]", s))
	case "PENDING", "WAITING", "WAITING_CONFIRMATION", "IN_PROGRESS", "UNPAID":
		return CAmb(fmt.Sprintf("[%s]", s))
	case "FAILED", "ERROR", "REJECTED", "EXPIRED", "NONAKTIF", "BATAL":
		return CDan(fmt.Sprintf("[%s]", s))
	default:
		return CDim(fmt.Sprintf("[%s]", s))
	}
}

func FormatMSISDN(val any) string {
	if val == nil {
		return "-"
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", val))
	if s == "" || s == "0" {
		return "-"
	}
	if strings.HasPrefix(s, "62") {
		s = "0" + s[2:]
	}
	if len(s) == 12 {
		return fmt.Sprintf("%s-%s-%s", s[:4], s[4:8], s[8:])
	} else if len(s) == 11 {
		return fmt.Sprintf("%s-%s-%s", s[:4], s[4:7], s[7:])
	} else if len(s) == 13 {
		return fmt.Sprintf("%s-%s-%s", s[:4], s[4:8], s[8:])
	}
	return s
}

func FormatRupiah(val any) string {
	if val == nil {
		return "Rp 0"
	}
	var num int64
	switch v := val.(type) {
	case int:
		num = int64(v)
	case int64:
		num = v
	case float64:
		num = int64(v)
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "Rp " + v
		}
		num = n
	default:
		return fmt.Sprintf("Rp %v", val)
	}

	str := strconv.FormatInt(num, 10)
	neg := false
	if num < 0 {
		neg = true
		str = str[1:]
	}

	var parts []string
	for len(str) > 3 {
		parts = append([]string{str[len(str)-3:]}, parts...)
		str = str[:len(str)-3]
	}
	if len(str) > 0 {
		parts = append([]string{str}, parts...)
	}

	res := strings.Join(parts, ".")
	if neg {
		return "Rp -" + res
	}
	return "Rp " + res
}

func FormatQuotaByte(quotaByte int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	if quotaByte >= TB {
		return fmt.Sprintf("%.2f TB", float64(quotaByte)/float64(TB))
	} else if quotaByte >= GB {
		return fmt.Sprintf("%.2f GB", float64(quotaByte)/float64(GB))
	} else if quotaByte >= MB {
		return fmt.Sprintf("%.2f MB", float64(quotaByte)/float64(MB))
	} else if quotaByte >= KB {
		return fmt.Sprintf("%.2f KB", float64(quotaByte)/float64(KB))
	}
	return fmt.Sprintf("%d B", quotaByte)
}

func RenderProgress(current, total int64, width int) string {
	if total <= 0 {
		return CSuc("[" + strings.Repeat("█", width) + "]")
	}
	pct := math.Min(1.0, math.Max(0.0, float64(current)/float64(total)))
	filled := int(pct * float64(width))
	empty := width - filled
	return CSuc("["+strings.Repeat("█", filled)) + CDim(strings.Repeat("░", empty)+"]")
}
