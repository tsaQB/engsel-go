package menu

import (
	"fmt"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/client/purchase"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func PurchaseNTimesByOptionCode(optionCode string, count int, useDecoy bool, pauseOnSuccess bool, delaySec int) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	tui.RenderAlert("Mengambil data paket...", "info")
	pkg, err := client.GetPackage(idToken, optionCode)
	if err != nil || pkg == nil {
		tui.RenderAlert("Gagal mengambil data paket.", "error")
		tui.TuiPause()
		return
	}

	pkgOpt, _ := pkg["package_option"].(map[string]any)
	tokenConfirmation, _ := pkg["token_confirmation"].(string)
	price := 0
	if pkgOpt != nil {
		switch v := pkgOpt["price"].(type) {
		case float64:
			price = int(v)
		case int:
			price = v
		}
	}

	paymentFor := "BUY_PACKAGE"
	if pf, ok := pkg["payment_for"].(string); ok && pf != "" {
		paymentFor = pf
	}

	var items []purchase.PaymentItem

	if useDecoy {
		decoy := service.DecoyInstance.GetDecoy("balance")
		if decoy != nil && decoy.OptionCode != "" {
			items = append(items, purchase.PaymentItem{
				ItemCode:          decoy.OptionCode,
				ProductType:       "",
				ItemPrice:         decoy.Price,
				ItemName:          "Decoy Package",
				Tax:               0,
				TokenConfirmation: tokenConfirmation,
			})
		}
	}

	items = append(items, purchase.PaymentItem{
		ItemCode:          optionCode,
		ProductType:       "",
		ItemPrice:         price,
		ItemName:          "Target Package",
		Tax:               0,
		TokenConfirmation: tokenConfirmation,
	})

	for i := 1; i <= count; i++ {
		tui.RenderAlert(fmt.Sprintf("[%d/%d] Memproses pembelian paket...", i, count), "info")
		res, err := purchase.SettlementBalance(activeUser.Tokens, items, paymentFor, price, len(items)-1, "", "")
		if err != nil {
			tui.RenderAlert(fmt.Sprintf("[%d/%d] Pembelian gagal: %v", i, count, err), "error")
		} else {
			status, _ := res["status"].(string)
			if status == "SUCCESS" {
				tui.RenderAlert(fmt.Sprintf("[%d/%d] Pembelian SUKSES!", i, count), "success")
				if pauseOnSuccess && i < count {
					tui.TuiPause("Tekan Enter untuk lanjut ke pembelian berikutnya...")
				}
			} else {
				tui.RenderAlert(fmt.Sprintf("[%d/%d] Status: %s", i, count, status), "warning")
			}
		}

		if delaySec > 0 && i < count {
			time.Sleep(time.Duration(delaySec) * time.Second)
		}
	}

	tui.RenderAlert("Selesai memproses seluruh perulangan pembelian.", "info")
	tui.TuiPause()
}

func PurchaseByFamily(familyCode string, useDecoy bool, pauseOnSuccess bool, delaySec int, startFromOption int) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	tui.RenderAlert("Mengambil daftar opsi dari family...", "info")
	familyData, err := client.GetFamily(idToken, familyCode, false)
	if err != nil || familyData == nil {
		tui.RenderAlert("Gagal memuat data family.", "error")
		tui.TuiPause()
		return
	}

	variants, _ := familyData["package_variants"].([]any)
	var optionCodes []string
	for _, v := range variants {
		vMap, _ := v.(map[string]any)
		if vMap == nil {
			continue
		}
		opts, _ := vMap["package_options"].([]any)
		for _, opt := range opts {
			optMap, _ := opt.(map[string]any)
			if optMap == nil {
				continue
			}
			code, _ := optMap["package_option_code"].(string)
			if code != "" {
				optionCodes = append(optionCodes, code)
			}
		}
	}

	startIdx := startFromOption - 1
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(optionCodes) {
		tui.RenderAlert("Nomor opsi mulai melebihi total opsi yang ada.", "warning")
		tui.TuiPause()
		return
	}

	for i := startIdx; i < len(optionCodes); i++ {
		optCode := optionCodes[i]
		tui.RenderAlert(fmt.Sprintf("[%d/%d] Membeli opsi: %s...", i+1, len(optionCodes), optCode), "info")
		PurchaseNTimesByOptionCode(optCode, 1, useDecoy, pauseOnSuccess, delaySec)
	}
}
