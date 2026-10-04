package menu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

type HotPackageItem struct {
	FamilyName   string `json:"family_name"`
	FamilyCode   string `json:"family_code"`
	IsEnterprise bool   `json:"is_enterprise"`
	VariantName  string `json:"variant_name"`
	OptionName   string `json:"option_name"`
	Order        int    `json:"order"`
}

func showHotList(title, filename string) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	path := filepath.Join("hot_data", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		tui.RenderAlert(fmt.Sprintf("Gagal memuat file %s: %v", path, err), "error")
		tui.TuiPause()
		return
	}

	var hotPackages []HotPackageItem
	if err := json.Unmarshal(data, &hotPackages); err != nil {
		tui.RenderAlert(fmt.Sprintf("Format JSON di %s tidak valid: %v", path, err), "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		tui.RenderHeader(title, "engsel ❯ Promo", tui.Width)
		fmt.Println(tui.RenderHdrLine("DAFTAR REKOMENDASI PAKET HOT", tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		for idx, p := range hotPackages {
			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			fullTitle := fmt.Sprintf("%s - %s - %s", p.FamilyName, p.VariantName, p.OptionName)
			titleClean := tui.TruncateStr(fullTitle, tui.Width-12)

			content := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(titleClean)
			plainLen := len(fmt.Sprintf("  [%2d] ", idx+1)) + len(titleClean)
			fmt.Println(tui.PLine(content, plainLen))
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
		if err == nil && idx >= 1 && idx <= len(hotPackages) {
			sel := hotPackages[idx-1]
			tui.ClearScreen()
			tui.RenderAlert("Mengambil data detail paket dari server...", "info")

			familyData, err := client.GetFamily(idToken, sel.FamilyCode, sel.IsEnterprise)
			if err != nil || familyData == nil {
				tui.RenderAlert("Gagal mengambil data family paket.", "error")
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
				if vName == sel.VariantName {
					opts, _ := vMap["package_options"].([]any)
					for _, opt := range opts {
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
						if orderVal == sel.Order {
							optCode, _ = optMap["package_option_code"].(string)
							break
						}
					}
				}
			}

			if optCode != "" {
				ShowPackageDetails(optCode, sel.IsEnterprise, sel.Order)
			} else {
				tui.RenderAlert("Kode opsi paket tidak ditemukan.", "error")
				tui.TuiPause()
			}
		} else {
			tui.RenderAlert("Nomor paket tidak valid.", "warning")
			tui.TuiPause()
		}
	}
}

func ShowHotMenu() {
	showHotList("PAKET HOT", "hot.json")
}

func ShowHotMenu2() {
	showHotList("PAKET HOT 2", "hot2.json")
}
