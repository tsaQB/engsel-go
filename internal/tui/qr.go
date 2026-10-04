package tui

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

// PrintQRCode renders a QR code in the terminal using Unicode blocks.
func PrintQRCode(content string) string {
	qr, err := qrcode.New(content, qrcode.Low)
	if err != nil {
		fmt.Printf("Error generating QR code: %v\n", err)
		return ""
	}

	bm := qr.Bitmap()
	border := 2
	size := len(bm)

	// Half-block terminal renderer for crisp, compact QR display
	// '▀' top half, '▄' bottom half, '█' full, ' ' empty
	fmt.Println()
	for y := -border; y < size+border; y += 2 {
		var row strings.Builder
		row.WriteString("  ")
		for x := -border; x < size+border; x++ {
			top := false
			if y >= 0 && y < size && x >= 0 && x < size {
				top = bm[y][x]
			}
			bot := false
			if y+1 >= 0 && y+1 < size && x >= 0 && x < size {
				bot = bm[y+1][x]
			}

			if top && bot {
				row.WriteString(" ")
			} else if top && !bot {
				row.WriteString("▄")
			} else if !top && bot {
				row.WriteString("▀")
			} else {
				row.WriteString("█")
			}
		}
		fmt.Println(row.String())
	}
	fmt.Println()

	qrisB64 := base64.URLEncoding.EncodeToString([]byte(content))
	qrisURL := fmt.Sprintf("https://ki-ar-kod.netlify.app/?data=%s", qrisB64)
	fmt.Println(TextDim("  Atau buka link berikut untuk melihat QRIS:"))
	fmt.Println("  " + Primary(qrisURL))
	fmt.Println()

	return qrisB64
}
