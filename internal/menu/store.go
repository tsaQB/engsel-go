package menu

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tsaQB/engsel-go/internal/client/store"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowStoreSegmentsMenu(isEnterprise bool) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil data Store Segments dari server...", "info")

		res, err := store.GetSegments(idToken, isEnterprise)
		if err != nil || res == nil {
			tui.RenderAlert("Tidak ada segmen promo ditemukan.", "warning")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		segments, _ := dataMap["store_segments"].([]any)
		if len(segments) == 0 {
			tui.RenderAlert("Daftar segmen promo kosong.", "info")
			tui.TuiPause()
			return
		}

		tui.ClearScreen()
		tui.RenderHeader("STORE SEGMENTS & PROMO", "engsel ❯ Promo ❯ Segments", tui.Width)
		fmt.Println(tui.RenderHdrLine(fmt.Sprintf("KATEGORI PROMO (%d)", len(segments)), tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		type SegmentAction struct {
			ActionParam string
			ActionType  string
		}
		packages := make(map[string]SegmentAction)

		for i, seg := range segments {
			segMap, _ := seg.(map[string]any)
			if segMap == nil {
				continue
			}
			name, _ := segMap["title"].(string)
			banners, _ := segMap["banners"].([]any)
			letter := string(rune('A' + i))

			fmt.Println(tui.RenderBoxDivider(fmt.Sprintf("BANNER %s: %s", letter, name), tui.Width))
			for j, b := range banners {
				bMap, _ := b.(map[string]any)
				if bMap == nil {
					continue
				}

				codeKey := fmt.Sprintf("%s%d", letter, j+1)
				actParam, _ := bMap["action_param"].(string)
				actType, _ := bMap["action_type"].(string)
				packages[strings.ToLower(codeKey)] = SegmentAction{
					ActionParam: actParam,
					ActionType:  actType,
				}

				priceStr := tui.FormatRupiah(bMap["discounted_price"])
				title, _ := bMap["title"].(string)
				famName, _ := bMap["family_name"].(string)
				validity, _ := bMap["validity"].(string)

				fullT := strings.Trim(fmt.Sprintf("%s - %s", famName, title), " -")
				availTW := (tui.Width - 2) - len(fmt.Sprintf("  [%s] ", codeKey)) - len(priceStr) - 2
				tClean := tui.TruncateStr(fullT, availTW)

				leftP := fmt.Sprintf("  [%s] %s", codeKey, tClean)
				rem := (tui.Width - 2) - len(leftP) - len(priceStr) - 1
				if rem < 1 {
					rem = 1
				}

				row1 := tui.BgSpaces(2) + tui.Badge(codeKey, tui.CPri) + tui.BgSpaces(1) + tui.CWhi(tClean) + tui.BgSpaces(rem) + tui.CAmb(priceStr) + tui.BgSpaces(1)
				fmt.Println(tui.PLine(row1, len(leftP)+rem+len(priceStr)+1))

				row2P := fmt.Sprintf("      Masa Aktif: %s", validity)
				row2 := tui.BgSpaces(6) + tui.CDim(row2P)
				fmt.Println(tui.PLine(row2, len(row2P)))
			}
			fmt.Println(tui.RenderEmptyLine())
		}

		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik kode paket (contoh: A1, B2)", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Kode promo (contoh: A1)")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		}

		selectedPkg, ok := packages[strings.ToLower(choice)]
		if !ok {
			tui.RenderAlert("Kode paket tidak valid. Masukkan kode seperti A1, B2.", "warning")
			tui.TuiPause()
			continue
		}

		if selectedPkg.ActionType == "PDP" {
			ShowPackageDetails(selectedPkg.ActionParam, isEnterprise)
		} else {
			tui.RenderAlert(fmt.Sprintf("Tipe aksi tidak didukung: %s (%s)", selectedPkg.ActionType, selectedPkg.ActionParam), "info")
			tui.TuiPause()
		}
	}
}

func ShowFamilyListMenu(subsType string, isEnterprise bool) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil daftar Family Paket dari Store...", "info")

		res, err := store.GetFamilyList(idToken, subsType, isEnterprise)
		if err != nil || res == nil {
			tui.RenderAlert("Tidak ada daftar family ditemukan.", "warning")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		familyList, _ := dataMap["results"].([]any)
		if len(familyList) == 0 {
			tui.RenderAlert("Daftar family paket kosong.", "info")
			tui.TuiPause()
			return
		}

		tui.ClearScreen()
		tui.RenderHeader("DAFTAR FAMILY PAKET", "engsel ❯ Store ❯ Family", tui.Width)
		fmt.Println(tui.RenderHdrLine(fmt.Sprintf("KATALOG FAMILY (%d)", len(familyList)), tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		for idx, item := range familyList {
			fMap, _ := item.(map[string]any)
			if fMap == nil {
				continue
			}
			fName, _ := fMap["label"].(string)
			fCode, _ := fMap["id"].(string)

			availW := (tui.Width - 2) - len(fmt.Sprintf("  [%2d] ", idx+1)) - 2
			nameClean := tui.TruncateStr(fName, availW)

			row1P := fmt.Sprintf("  [%2d] %s", idx+1, nameClean)
			row1 := tui.BgSpaces(2) + tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri) + tui.BgSpaces(1) + tui.CWhi(nameClean)
			fmt.Println(tui.PLine(row1, len(row1P)))

			row2P := fmt.Sprintf("      Kode: %s", fCode)
			row2 := tui.BgSpaces(6) + tui.CDim(row2P)
			fmt.Println(tui.PLine(row2, len(row2P)))

			if idx < len(familyList)-1 {
				fmt.Println(tui.RenderEmptyLine())
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik nomor family (1..N)", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilih family (nomor)")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		}

		idx, err := strconv.Atoi(choice)
		if err == nil && idx >= 1 && idx <= len(familyList) {
			target, _ := familyList[idx-1].(map[string]any)
			fCode, _ := target["id"].(string)
			GetPackagesByFamily(fCode, isEnterprise)
		} else {
			tui.RenderAlert("Nomor pilihan tidak valid.", "warning")
			tui.TuiPause()
		}
	}
}

func ShowStorePackagesMenu(subsType string, isEnterprise bool) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil paket dari Store Search...", "info")

		res, err := store.GetStorePackages(idToken, subsType, isEnterprise)
		if err != nil || res == nil {
			tui.RenderAlert("Tidak ada paket store ditemukan.", "warning")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		pkgs, _ := dataMap["packages"].([]any)
		if len(pkgs) == 0 {
			tui.RenderAlert("Daftar paket store kosong.", "info")
			tui.TuiPause()
			return
		}

		tui.ClearScreen()
		tui.RenderHeader("KATALOG STORE PACKAGES", "engsel ❯ Store ❯ Packages", tui.Width)
		fmt.Println(tui.RenderHdrLine(fmt.Sprintf("DAFTAR PAKET (%d)", len(pkgs)), tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		for idx, item := range pkgs {
			pMap, _ := item.(map[string]any)
			if pMap == nil {
				continue
			}
			pName, _ := pMap["name"].(string)
			priceStr := tui.FormatRupiah(pMap["price"])
			validity, _ := pMap["validity"].(string)

			availW := (tui.Width - 2) - len(fmt.Sprintf("  [%2d] ", idx+1)) - len(priceStr) - 2
			nameClean := tui.TruncateStr(pName, availW)

			leftP := fmt.Sprintf("  [%2d] %s", idx+1, nameClean)
			rem := (tui.Width - 2) - len(leftP) - len(priceStr) - 1
			if rem < 1 {
				rem = 1
			}

			row1 := tui.BgSpaces(2) + tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri) + tui.BgSpaces(1) + tui.CWhi(nameClean) + tui.BgSpaces(rem) + tui.CAmb(priceStr) + tui.BgSpaces(1)
			fmt.Println(tui.PLine(row1, len(leftP)+rem+len(priceStr)+1))

			row2P := fmt.Sprintf("      Masa Aktif: %s", validity)
			row2 := tui.BgSpaces(6) + tui.CDim(row2P)
			fmt.Println(tui.PLine(row2, len(row2P)))

			if idx < len(pkgs)-1 {
				fmt.Println(tui.RenderEmptyLine())
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik nomor paket (1..N)", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilih paket (nomor)")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		}

		idx, err := strconv.Atoi(choice)
		if err == nil && idx >= 1 && idx <= len(pkgs) {
			target, _ := pkgs[idx-1].(map[string]any)
			optCode, _ := target["package_option_code"].(string)
			ShowPackageDetails(optCode, isEnterprise)
		} else {
			tui.RenderAlert("Nomor paket tidak valid.", "warning")
			tui.TuiPause()
		}
	}
}

func ShowRedeemablesMenu(isEnterprise bool) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil data Redeemables dari server...", "info")

		res, err := store.GetRedeemables(idToken, isEnterprise)
		if err != nil || res == nil {
			tui.RenderAlert("Tidak ada redeemables ditemukan.", "warning")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		categories, _ := dataMap["categories"].([]any)
		if len(categories) == 0 {
			tui.RenderAlert("Daftar kategori redeemables kosong.", "info")
			tui.TuiPause()
			return
		}

		tui.ClearScreen()
		tui.RenderHeader("KATALOG REDEEMABLES", "engsel ❯ Store ❯ Redeem", tui.Width)
		fmt.Println(tui.RenderHdrLine(fmt.Sprintf("KATEGORI REDEEMABLES (%d)", len(categories)), tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		type RedeemAction struct {
			ActionParam string
			ActionType  string
		}
		packages := make(map[string]RedeemAction)

		for i, cat := range categories {
			catMap, _ := cat.(map[string]any)
			if catMap == nil {
				continue
			}
			catName, _ := catMap["category_name"].(string)
			redeems, _ := catMap["redeemables"].([]any)
			letter := string(rune('A' + i))

			fmt.Println(tui.RenderBoxDivider(fmt.Sprintf("KATEGORI %s: %s", letter, catName), tui.Width))
			if len(redeems) == 0 {
				emptyRow := "  Tidak ada item redeemable di kategori ini."
				fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(emptyRow[2:]), len(emptyRow)))
			} else {
				for j, red := range redeems {
					rMap, _ := red.(map[string]any)
					if rMap == nil {
						continue
					}
					name, _ := rMap["name"].(string)
					codeKey := fmt.Sprintf("%s%d", letter, j+1)
					actParam, _ := rMap["action_param"].(string)
					actType, _ := rMap["action_type"].(string)
					packages[strings.ToLower(codeKey)] = RedeemAction{
						ActionParam: actParam,
						ActionType:  actType,
					}

					var validUntil int64
					switch v := rMap["valid_until"].(type) {
					case float64:
						validUntil = int64(v)
					case int64:
						validUntil = v
					}
					dateStr := "-"
					if validUntil > 0 {
						dateStr = time.Unix(validUntil, 0).Format("02 Jan 2006")
					}

					availW := (tui.Width - 2) - len(fmt.Sprintf("  [%s] ", codeKey)) - 2
					nameClean := tui.TruncateStr(name, availW)

					row1P := fmt.Sprintf("  [%s] %s", codeKey, nameClean)
					row1 := tui.BgSpaces(2) + tui.Badge(codeKey, tui.CPri) + tui.BgSpaces(1) + tui.CWhi(nameClean)
					fmt.Println(tui.PLine(row1, len(row1P)))

					row2P := fmt.Sprintf("      Berlaku s/d: %s", dateStr)
					row2 := tui.BgSpaces(6) + tui.CDim(row2P)
					fmt.Println(tui.PLine(row2, len(row2P)))
				}
			}
			fmt.Println(tui.RenderEmptyLine())
		}

		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik kode item (contoh: A1, B2)", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Kode item (contoh: A1)")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		}

		selected, ok := packages[strings.ToLower(choice)]
		if !ok {
			tui.RenderAlert("Kode item tidak valid. Masukkan kode seperti A1, B2.", "warning")
			tui.TuiPause()
			continue
		}

		if selected.ActionType == "PLP" {
			GetPackagesByFamily(selected.ActionParam, isEnterprise)
		} else if selected.ActionType == "PDP" {
			ShowPackageDetails(selected.ActionParam, isEnterprise)
		} else {
			tui.RenderAlert(fmt.Sprintf("Tipe aksi tidak didukung: %s (%s)", selected.ActionType, selected.ActionParam), "info")
			tui.TuiPause()
		}
	}
}
