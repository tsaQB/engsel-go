package menu

import (
	"fmt"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowTransactionHistory() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		tui.RenderHeader("RIWAYAT TRANSAKSI", "engsel ❯ Transaksi", tui.Width)

		idToken, _ := activeUser.Tokens["id_token"].(string)
		data, err := client.GetTransactionHistory(idToken)
		var history []any
		if err == nil && data != nil {
			if l, ok := data["list"].([]any); ok {
				history = l
			}
		}

		fmt.Println(tui.RenderHdrLine("DAFTAR TRANSAKSI TERAKHIR", tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		if len(history) == 0 {
			emptyStr := "  Tidak ada data riwayat transaksi."
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(emptyStr[2:]), len(emptyStr)))
		} else {
			for idx, item := range history {
				tx, ok := item.(map[string]any)
				if !ok {
					continue
				}

				var ts int64
				switch v := tx["timestamp"].(type) {
				case float64:
					ts = int64(v)
				case int64:
					ts = v
				}

				formattedTime := "-"
				if ts > 0 {
					formattedTime = time.Unix(ts, 0).Format("02 Jan 2006, 15:04 WIB")
				}

				title, _ := tx["title"].(string)
				if title == "" {
					title = "Transaksi"
				}
				priceStr := tui.FormatRupiah(tx["price"])
				payMethod, _ := tx["payment_method_label"].(string)
				status, _ := tx["status"].(string)
				payStatus, _ := tx["payment_status"].(string)

				// Row 1: [Idx] Title  Price
				kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
				pLen := len(priceStr)
				availW := (tui.Width - 2) - len(fmt.Sprintf("  [%2d] ", idx+1)) - pLen - 2
				titleClean := tui.TruncateStr(title, availW)

				leftP := fmt.Sprintf("  [%2d] %s", idx+1, titleClean)
				rem := (tui.Width - 2) - len(leftP) - pLen - 1
				if rem < 1 {
					rem = 1
				}
				row1 := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(titleClean) + tui.BgSpaces(rem) + tui.CAmb(priceStr) + tui.BgSpaces(1)
				fmt.Println(tui.PLine(row1, len(leftP)+rem+pLen+1))

				// Row 2: Time & Method
				line2Plain := fmt.Sprintf("      %s • %s", formattedTime, payMethod)
				line2Clean := tui.TruncateStr(line2Plain, tui.Width-4)
				row2 := tui.BgSpaces(6) + tui.CDim(line2Clean[6:])
				fmt.Println(tui.PLine(row2, len(line2Clean)))

				// Row 3: Status
				statBadge := tui.StatusBadge(status)
				payStatBadge := tui.StatusBadge(payStatus)
				plainStat := fmt.Sprintf("      Status: %s | Pembayaran: %s", status, payStatus)
				row3 := tui.BgSpaces(6) + tui.CMut("Status: ") + statBadge + tui.CMut(" | ") + payStatBadge
				fmt.Println(tui.PLine(row3, len(plainStat)))

				if idx < len(history)-1 {
					fmt.Println(tui.RenderEmptyLine())
				}
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))

		lRef := "  [ 0] Refresh Data Transaksi"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("0", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Refresh Data Transaksi"), len(lRef)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan riwayat")
		if choice == "0" {
			continue
		} else if choice == "00" || choice == "q" || choice == "exit" {
			return
		} else {
			tui.RenderAlert("Opsi tidak valid. Silakan pilih 0 atau 00.", "warning")
			tui.TuiPause()
		}
	}
}
