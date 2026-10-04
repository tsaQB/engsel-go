package menu

import (
	"fmt"
	"strconv"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowBookmarkMenu() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		bookmarks := service.BookmarkInstance.GetBookmarks()

		tui.RenderHeader("BOOKMARK PAKET", "engsel ❯ Bookmark", tui.Width)
		fmt.Println(tui.RenderHdrLine("DAFTAR PAKET TERSIMPAN", tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		if len(bookmarks) == 0 {
			emptyStr := "  Tidak ada paket yang disimpan di bookmark."
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(emptyStr[2:]), len(emptyStr)))
			fmt.Println(tui.RenderEmptyLine())
			fmt.Println(tui.RenderBoxBottom("", tui.Width))
			tui.TuiPause()
			return
		}

		for idx, bm := range bookmarks {
			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			fullTitle := fmt.Sprintf("%s - %s - %s", bm.FamilyName, bm.VariantName, bm.OptionName)
			titleTrunc := tui.TruncateStr(fullTitle, tui.Width-12)

			content := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(titleTrunc)
			plainLen := len(fmt.Sprintf("  [%2d] ", idx+1)) + len(titleTrunc)
			fmt.Println(tui.PLine(content, plainLen))
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("PILIHAN AKSI", tui.Width))

		lSelect := "  [1..N] Ketik nomor paket untuk melihat detail & beli"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1..N", tui.CPri)+tui.BgSpaces(1)+tui.CMut("Ketik nomor paket untuk melihat detail & beli"), len(lSelect)))

		lDel := "  [000] Hapus Paket dari Bookmark"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("000", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Hapus Paket dari Bookmark"), len(lDel)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik nomor/opsi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan bookmark")
		if choice == "00" || choice == "q" || choice == "exit" {
			return
		} else if choice == "000" {
			delChoice := tui.TuiInput("Nomor bookmark yang ingin dihapus")
			idx, err := strconv.Atoi(delChoice)
			if err == nil && idx >= 1 && idx <= len(bookmarks) {
				delBM := bookmarks[idx-1]
				titleDel := fmt.Sprintf("%s - %s", delBM.FamilyName, delBM.OptionName)
				if tui.TuiConfirm(fmt.Sprintf("Hapus bookmark '%s'?", titleDel), true) {
					service.BookmarkInstance.RemoveBookmark(delBM.FamilyCode, delBM.IsEnterprise, delBM.VariantName, delBM.Order)
					tui.RenderAlert("Bookmark berhasil dihapus.", "success")
				} else {
					tui.RenderAlert("Penghapusan bookmark dibatalkan.", "info")
				}
				tui.TuiPause()
			} else {
				tui.RenderAlert("Nomor bookmark tidak valid.", "error")
				tui.TuiPause()
			}
		} else {
			idx, err := strconv.Atoi(choice)
			if err == nil && idx >= 1 && idx <= len(bookmarks) {
				selBM := bookmarks[idx-1]
				tui.ClearScreen()
				tui.RenderAlert("Mengambil data paket dari server MyXL...", "info")

				idToken, _ := activeUser.Tokens["id_token"].(string)
				familyData, err := client.GetFamily(idToken, selBM.FamilyCode, selBM.IsEnterprise)
				if err != nil || familyData == nil {
					tui.RenderAlert("Gagal mengambil data family paket dari server.", "error")
					tui.TuiPause()
					continue
				}

				variants, _ := familyData["package_variants"].([]any)
				optCode := ""
				for _, v := range variants {
					vMap, _ := v.(map[string]any)
					if vMap == nil {
						continue
					}
					vName, _ := vMap["name"].(string)
					if vName == selBM.VariantName {
						options, _ := vMap["package_options"].([]any)
						for _, opt := range options {
							optMap, _ := opt.(map[string]any)
							if optMap == nil {
								continue
							}
							orderVal := 0
							switch ov := optMap["order"].(type) {
							case float64:
								orderVal = int(ov)
							case int:
								orderVal = ov
							}
							if orderVal == selBM.Order {
								optCode, _ = optMap["package_option_code"].(string)
								break
							}
						}
					}
				}

				if optCode != "" {
					ShowPackageDetails(optCode, selBM.IsEnterprise)
				} else {
					tui.RenderAlert("Kode paket tidak ditemukan atau paket sudah tidak tersedia.", "error")
					tui.TuiPause()
				}
			} else {
				tui.RenderAlert("Input tidak valid. Silakan coba lagi.", "error")
				tui.TuiPause()
			}
		}
	}
}
