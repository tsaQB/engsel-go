package menu

import (
	"fmt"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowNotificationMenu() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		tui.RenderHeader("PUSAT NOTIFIKASI", "engsel ❯ Notifikasi", tui.Width)

		idToken, _ := activeUser.Tokens["id_token"].(string)
		res, err := client.DashboardSegments(idToken)
		if err != nil || res == nil {
			tui.RenderAlert("Tidak ada data notifikasi ditemukan dari server.", "info")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		notifMap, _ := dataMap["notification"].(map[string]any)
		notifications, _ := notifMap["data"].([]any)
		if len(notifications) == 0 {
			tui.RenderAlert("Kotak masuk notifikasi Anda kosong.", "info")
			tui.TuiPause()
			return
		}

		unreadCount := 0
		for _, n := range notifications {
			nMap, _ := n.(map[string]any)
			isRead, _ := nMap["is_read"].(bool)
			if !isRead {
				unreadCount++
			}
		}

		summary := fmt.Sprintf("Total: %d Pesan | Belum Dibaca: %d", len(notifications), unreadCount)
		fmt.Println(tui.RenderHdrLine(summary, tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		for idx, n := range notifications {
			nMap, _ := n.(map[string]any)
			isRead, _ := nMap["is_read"].(bool)
			brief, _ := nMap["brief_message"].(string)
			if brief == "" {
				brief = "Pemberitahuan"
			}
			full, _ := nMap["full_message"].(string)
			timeStr, _ := nMap["timestamp"].(string)

			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			statBadge := tui.CAmb("[BARU]")
			if isRead {
				statBadge = tui.CDim("[SUDAH DIBACA]")
			}

			availW := (tui.Width - 2) - len(fmt.Sprintf("  [%2d] ", idx+1)) - len(" [BARU] ") - 2
			briefTrunc := tui.TruncateStr(brief, availW)
			row1P := fmt.Sprintf("  [%2d] [BARU] %s", idx+1, briefTrunc)
			if isRead {
				row1P = fmt.Sprintf("  [%2d] [DIBACA] %s", idx+1, briefTrunc)
			}

			textFn := tui.CWhi
			if isRead {
				textFn = tui.CMut
			}

			content1 := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + statBadge + tui.BgSpaces(1) + textFn(briefTrunc)
			fmt.Println(tui.PLine(content1, len(row1P)))

			row2P := fmt.Sprintf("      Waktu: %s", timeStr)
			content2 := tui.BgSpaces(6) + tui.CDim(fmt.Sprintf("Waktu: %s", timeStr))
			fmt.Println(tui.PLine(content2, len(row2P)))

			if full != "" && full != brief {
				fullTrunc := tui.TruncateStr(full, tui.Width-10)
				row3P := fmt.Sprintf("      %s", fullTrunc)
				content3 := tui.BgSpaces(6) + tui.CMut(fullTrunc)
				fmt.Println(tui.PLine(content3, len(row3P)))
			}

			if idx < len(notifications)-1 {
				fmt.Println(tui.RenderEmptyLine())
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))

		lRead := "  [ 1] Tandai Semua Notifikasi Sudah Dibaca"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CSuc)+tui.BgSpaces(1)+tui.CWhi("Tandai Semua Notifikasi Sudah Dibaca"), len(lRead)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan notifikasi")
		if choice == "1" {
			marked := 0
			for _, n := range notifications {
				nMap, _ := n.(map[string]any)
				isRead, _ := nMap["is_read"].(bool)
				if isRead {
					continue
				}
				nID, _ := nMap["notification_id"].(string)
				_, err := client.GetNotificationDetail(idToken, nID)
				if err == nil {
					marked++
				}
			}
			tui.RenderAlert(fmt.Sprintf("%d notifikasi berhasil ditandai sebagai sudah dibaca.", marked), "success")
			tui.TuiPause()
		} else if choice == "00" || choice == "q" || choice == "exit" {
			return
		} else {
			tui.RenderAlert("Pilihan tidak valid. Silakan coba lagi.", "warning")
			tui.TuiPause()
		}
	}
}
