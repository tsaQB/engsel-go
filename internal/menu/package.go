package menu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/client/purchase"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowPackageDetails(packageOptionCode string, isEnterprise bool, optionOrder ...int) bool {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return false
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	tui.ClearScreen()
	tui.RenderAlert("Mengambil data detail paket dari server MyXL...", "info")

	pkg, err := client.GetPackage(idToken, packageOptionCode)
	if err != nil || pkg == nil {
		tui.RenderAlert("Gagal memuat data detail paket dari server.", "error")
		tui.TuiPause()
		return false
	}

	pkgOpt, _ := pkg["package_option"].(map[string]any)
	pkgFam, _ := pkg["package_family"].(map[string]any)
	pkgVar, _ := pkg["package_detail_variant"].(map[string]any)
	pkgAddon, _ := pkg["package_addon"].(map[string]any)

	price := 0
	if pkgOpt != nil {
		switch v := pkgOpt["price"].(type) {
		case float64:
			price = int(v)
		case int:
			price = v
		}
	}

	validity := "-"
	optName := ""
	pointsVal := 0
	if pkgOpt != nil {
		if v, ok := pkgOpt["validity"].(string); ok && v != "" {
			validity = v
		}
		optName, _ = pkgOpt["name"].(string)
		switch v := pkgOpt["point"].(type) {
		case float64:
			pointsVal = int(v)
		case int:
			pointsVal = v
		}
	}

	famName := ""
	famCode := ""
	paymentFor := "BUY_PACKAGE"
	planType := "PREPAID"
	if pkgFam != nil {
		famName, _ = pkgFam["name"].(string)
		famCode, _ = pkgFam["package_family_code"].(string)
		if pf, ok := pkgFam["payment_for"].(string); ok && pf != "" {
			paymentFor = pf
		}
		if pt, ok := pkgFam["plan_type"].(string); ok && pt != "" {
			planType = pt
		}
	}

	varName := ""
	if pkgVar != nil {
		varName, _ = pkgVar["name"].(string)
	}

	parentCode := "N/A"
	if pkgAddon != nil {
		if pc, ok := pkgAddon["parent_code"].(string); ok && pc != "" {
			parentCode = pc
		}
	}

	tokenConfirmation, _ := pkg["token_confirmation"].(string)

	title := strings.TrimSpace(fmt.Sprintf("%s - %s - %s", famName, varName, optName))

	paymentItems := []purchase.PaymentItem{
		{
			ItemCode:          packageOptionCode,
			ProductType:       "",
			ItemPrice:         price,
			ItemName:          strings.TrimSpace(fmt.Sprintf("%s %s", varName, optName)),
			Tax:               0,
			TokenConfirmation: tokenConfirmation,
		},
	}

	for {
		tui.ClearScreen()
		tui.RenderHeader("DETAIL PAKET", "engsel ❯ Paket ❯ Detail", tui.Width)

		fmt.Println(tui.RenderHdrLine("INFORMASI PAKET", tui.Width))
		fmt.Println(tui.RenderKVLine("Nama Paket", title, tui.CWhi, tui.Width))
		fmt.Println(tui.RenderKV2Col("Harga", tui.FormatRupiah(price), tui.CAmb, "Masa Aktif", validity, tui.CWhi, 28, tui.Width))
		fmt.Println(tui.RenderKV2Col("Poin", fmt.Sprintf("%d Pts", pointsVal), tui.CPri, "Plan Type", planType, tui.CMut, 28, tui.Width))
		fmt.Println(tui.RenderKV2Col("Family", famCode, tui.CDim, "Parent", parentCode, tui.CDim, 28, tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		// Benefits
		if pkgOpt != nil {
			if benefits, ok := pkgOpt["benefits"].([]any); ok && len(benefits) > 0 {
				fmt.Println(tui.RenderBoxDivider("BENEFIT & KUOTA", tui.Width))
				for _, b := range benefits {
					bMap, _ := b.(map[string]any)
					if bMap == nil {
						continue
					}
					bName, _ := bMap["name"].(string)
					dataType, _ := bMap["data_type"].(string)
					var total int64
					switch v := bMap["total"].(type) {
					case float64:
						total = int64(v)
					case int64:
						total = v
					case int:
						total = int64(v)
					}

					valStr := ""
					if dataType == "VOICE" && total > 0 {
						valStr = fmt.Sprintf("%d Menit", total/60)
					} else if dataType == "TEXT" && total > 0 {
						valStr = fmt.Sprintf("%d SMS", total)
					} else if dataType == "DATA" && total > 0 {
						valStr = tui.FormatQuotaByte(total)
					} else {
						valStr = fmt.Sprintf("%d", total)
					}

					fmt.Println(tui.RenderKVLine(bName, valStr, tui.CSuc, tui.Width))
				}
				fmt.Println(tui.RenderEmptyLine())
			}
		}

		fmt.Println(tui.RenderBoxDivider("METODE PEMBAYARAN & AKSI", tui.Width))

		l1 := "  [ 1] Beli dengan Pulsa (Balance)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CSuc)+tui.BgSpaces(1)+tui.CWhi("Beli dengan Pulsa (Balance)"), len(l1)))

		l2 := "  [ 2] Beli dengan QRIS"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("2", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Beli dengan QRIS"), len(l2)))

		l3 := "  [ 3] Beli dengan E-Wallet (DANA, ShopeePay, GoPay, OVO)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("3", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Beli dengan E-Wallet (DANA, ShopeePay, GoPay, OVO)"), len(l3)))

		l4 := "  [ 4] Beli Ulang (Loop / Bulk)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("4", tui.CAmb)+tui.BgSpaces(1)+tui.CWhi("Beli Ulang (Loop / Bulk)"), len(l4)))

		l5 := "  [ 5] Simpan Paket ke Bookmark"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("5", tui.CPur)+tui.BgSpaces(1)+tui.CWhi("Simpan Paket ke Bookmark"), len(l5)))

		l6 := "  [ 6] Redeem Bounty (Voucher)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("6", tui.CPri)+tui.BgSpaces(1)+tui.CMut("Redeem Bounty (Voucher)"), len(l6)))

		l7 := "  [ 7] Redeem Loyalty Points"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("7", tui.CPri)+tui.BgSpaces(1)+tui.CMut("Redeem Loyalty Points"), len(l7)))

		l8 := "  [ 8] Hadiahkan / Kirim Bounty ke Nomor Lain"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("8", tui.CPri)+tui.BgSpaces(1)+tui.CMut("Hadiahkan / Kirim Bounty ke Nomor Lain"), len(l8)))

		lBack := "  [00] Kembali"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih metode/aksi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan")
		if choice == "00" || choice == "q" || choice == "exit" {
			return true
		} else if choice == "1" {
			// Pulsa (Balance)
			amountToPay := price
			if tui.TuiConfirm(fmt.Sprintf("Konfirmasi beli paket '%s' seharga %s?", title, tui.FormatRupiah(amountToPay)), true) {
				tui.RenderAlert("Mengirim transaksi pembayaran pulsa...", "info")
				res, err := purchase.SettlementBalance(activeUser.Tokens, paymentItems, paymentFor, amountToPay, 0, "", "")
				if err != nil {
					tui.RenderAlert(fmt.Sprintf("Pembelian gagal: %v", err), "error")
				} else {
					status, _ := res["status"].(string)
					if status == "SUCCESS" {
						tui.RenderAlert("Pembelian paket BERHASIL!", "success")
					} else {
						tui.RenderAlert(fmt.Sprintf("Status: %s", status), "warning")
					}
				}
				tui.TuiPause()
			}
		} else if choice == "2" {
			// QRIS
			tui.RenderAlert("Membuat transaksi QRIS...", "info")
			txID, err := purchase.SettlementQRIS(activeUser.Tokens, paymentItems, paymentFor, price, 0, "", "")
			if err != nil || txID == "" {
				tui.RenderAlert(fmt.Sprintf("Gagal membuat transaksi QRIS: %v", err), "error")
				tui.TuiPause()
				continue
			}

			tui.RenderAlert("Mengambil kode QRIS dari server...", "info")
			qrData, err := purchase.GetQRISCode(idToken, txID)
			if err != nil || qrData == "" {
				tui.RenderAlert(fmt.Sprintf("Gagal mendapatkan kode QRIS: %v", err), "error")
				tui.TuiPause()
				continue
			}

			tui.ClearScreen()
			tui.RenderHeader("PEMBAYARAN QRIS", "engsel ❯ QRIS", tui.Width)
			tui.PrintQRCode(qrData)
			tui.TuiPause("Tekan Enter jika sudah menyelesaikan pembayaran...")
		} else if choice == "3" {
			// E-Wallet
			tui.ClearScreen()
			tui.RenderHeader("PILIHAN E-WALLET", "engsel ❯ E-Wallet", tui.Width)
			fmt.Println("  1. DANA")
			fmt.Println("  2. ShopeePay")
			fmt.Println("  3. GoPay")
			fmt.Println("  4. OVO")
			fmt.Println()
			ewChoice := tui.TuiInput("Pilih e-wallet (1-4)")
			walletMethod := ""
			walletNumber := ""

			switch ewChoice {
			case "1":
				walletMethod = "DANA"
				walletNumber = tui.TuiInput("Nomor DANA (contoh: 08123456789)")
			case "2":
				walletMethod = "SHOPEEPAY"
			case "3":
				walletMethod = "GOPAY"
			case "4":
				walletMethod = "OVO"
				walletNumber = tui.TuiInput("Nomor OVO (contoh: 08123456789)")
			default:
				tui.RenderAlert("Pilihan e-wallet tidak valid.", "error")
				tui.TuiPause()
				continue
			}

			tui.RenderAlert("Memproses multipayment...", "info")
			res, err := purchase.SettlementMultipayment(activeUser.Tokens, paymentItems, walletNumber, walletMethod, paymentFor, price, 0)
			if err != nil || res == nil {
				tui.RenderAlert(fmt.Sprintf("Pembayaran e-wallet gagal: %v", err), "error")
			} else {
				dataMap, _ := res["data"].(map[string]any)
				deeplink, _ := dataMap["deeplink"].(string)
				if deeplink != "" {
					fmt.Println()
					fmt.Println("  Selesaikan pembayaran melalui link berikut:")
					fmt.Println("  " + tui.Primary(deeplink))
				} else if walletMethod == "OVO" {
					tui.RenderAlert("Silakan buka aplikasi OVO untuk konfirmasi pembayaran.", "info")
				} else {
					tui.RenderAlert("Transaksi e-wallet berhasil dibuat.", "success")
				}
			}
			tui.TuiPause()
		} else if choice == "4" {
			// Loop / Bulk purchase
			countStr := tui.TuiInput("Berapa kali pembelian ingin diulang?", "1")
			count, err := strconv.Atoi(countStr)
			if err == nil && count > 0 {
				delayStr := tui.TuiInput("Jeda antar pembelian dalam detik (0 untuk tanpa jeda)", "0")
				delaySec, _ := strconv.Atoi(delayStr)
				useDecoy := tui.TuiConfirm("Gunakan paket decoy untuk bypass?", false)
				pauseOnSuccess := tui.TuiConfirm("Jeda (pause) setiap kali pembelian sukses?", true)
				PurchaseNTimesByOptionCode(packageOptionCode, count, useDecoy, pauseOnSuccess, delaySec)
			}
		} else if choice == "5" {
			// Bookmark
			ord := 1
			if len(optionOrder) > 0 {
				ord = optionOrder[0]
			}
			saved := service.BookmarkInstance.AddBookmark(famCode, famName, isEnterprise, varName, optName, ord)
			if saved {
				tui.RenderAlert("Paket berhasil disimpan ke bookmark!", "success")
			} else {
				tui.RenderAlert("Paket sudah ada di bookmark.", "info")
			}
			tui.TuiPause()
		} else if choice == "6" {
			// Redeem Bounty
			var tsToSign int64
			switch v := pkg["timestamp"].(type) {
			case float64:
				tsToSign = int64(v)
			case int64:
				tsToSign = v
			}
			tui.RenderAlert("Mengirim klaim bounty voucher...", "info")
			res, err := purchase.SettlementBounty(activeUser.Tokens, tokenConfirmation, tsToSign, packageOptionCode, price)
			if err != nil {
				tui.RenderAlert(fmt.Sprintf("Gagal klaim bounty: %v", err), "error")
			} else {
				tui.RenderAlert("Klaim bounty berhasil dikirim!", "success")
				_ = res
			}
			tui.TuiPause()
		} else if choice == "7" {
			// Redeem Loyalty
			var tsToSign int64
			switch v := pkg["timestamp"].(type) {
			case float64:
				tsToSign = int64(v)
			case int64:
				tsToSign = v
			}
			tui.RenderAlert("Mengirim klaim loyalty poin...", "info")
			res, err := purchase.SettlementLoyalty(activeUser.Tokens, tokenConfirmation, tsToSign, packageOptionCode, price)
			if err != nil {
				tui.RenderAlert(fmt.Sprintf("Gagal tukar loyalty poin: %v", err), "error")
			} else {
				tui.RenderAlert("Penukaran loyalty poin berhasil dikirim!", "success")
				_ = res
			}
			tui.TuiPause()
		} else if choice == "8" {
			// Hadiahkan / Allotment
			var tsToSign int64
			switch v := pkg["timestamp"].(type) {
			case float64:
				tsToSign = int64(v)
			case int64:
				tsToSign = v
			}
			destRaw := tui.TuiInput("Nomor tujuan pengiriman (contoh: 08123456789)")
			if destRaw != "" {
				destMsisdn := config.NormalizeMSISDN(destRaw)
				tui.RenderAlert(fmt.Sprintf("Mengirim bounty ke %s...", tui.FormatMSISDN(destMsisdn)), "info")
				res, err := purchase.BountyAllotment(activeUser.Tokens, tsToSign, destMsisdn, title, packageOptionCode, tokenConfirmation)
				if err != nil {
					tui.RenderAlert(fmt.Sprintf("Gagal kirim bounty: %v", err), "error")
				} else {
					tui.RenderAlert(fmt.Sprintf("Bounty berhasil dikirim ke %s!", tui.FormatMSISDN(destMsisdn)), "success")
					_ = res
				}
				tui.TuiPause()
			}
		}
	}
}

func FetchMyPackages() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	tui.ClearScreen()
	tui.RenderAlert("Mengambil data paket aktif dari server...", "info")

	data, err := client.GetPendingTransaction(idToken)
	if err != nil || data == nil {
		tui.RenderAlert("Gagal memuat paket aktif.", "error")
		tui.TuiPause()
		return
	}

	packages, _ := data["quotas"].([]any)

	tui.ClearScreen()
	tui.RenderHeader("PAKET AKTIF SAYA", "engsel ❯ Paket Saya", tui.Width)

	if len(packages) == 0 {
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut("Tidak ada paket aktif yang ditemukan."), len("  Tidak ada paket aktif yang ditemukan.")))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("", tui.Width))
		tui.TuiPause()
		return
	}

	fmt.Println(tui.RenderHdrLine(fmt.Sprintf("DAFTAR PAKET AKTIF (%d)", len(packages)), tui.Width))
	fmt.Println(tui.RenderEmptyLine())

	for idx, p := range packages {
		pMap, _ := p.(map[string]any)
		if pMap == nil {
			continue
		}

		name, _ := pMap["name"].(string)
		optCode, _ := pMap["package_option_code"].(string)
		expStr, _ := pMap["expired_at"].(string)

		var totalQuota int64
		var remainingQuota int64
		switch v := pMap["total"].(type) {
		case float64:
			totalQuota = int64(v)
		case int64:
			totalQuota = v
		}
		switch v := pMap["remaining"].(type) {
		case float64:
			remainingQuota = int64(v)
		case int64:
			remainingQuota = v
		}

		kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
		tTrunc := tui.TruncateStr(name, tui.Width-12)
		r1 := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(tTrunc)
		fmt.Println(tui.PLine(r1, len(fmt.Sprintf("  [%2d] ", idx+1))+len(tTrunc)))

		if totalQuota > 0 {
			qBar := tui.RenderProgress(remainingQuota, totalQuota, 16)
			remFmt := tui.FormatQuotaByte(remainingQuota)
			totFmt := tui.FormatQuotaByte(totalQuota)
			r2Str := fmt.Sprintf("      Kuota: %s / %s", remFmt, totFmt)
			r2 := tui.BgSpaces(6) + tui.CMut("Kuota: ") + tui.CAmb(remFmt) + tui.CMut(" / ") + tui.CWhi(totFmt) + tui.BgSpaces(2) + qBar
			fmt.Println(tui.PLine(r2, len(r2Str)+2+18))
		}

		r3Str := fmt.Sprintf("      Berlaku s/d: %s • Kode: %s", expStr, optCode)
		r3Clean := tui.TruncateStr(r3Str, tui.Width-4)
		r3 := tui.BgSpaces(6) + tui.CDim(r3Clean[6:])
		fmt.Println(tui.PLine(r3, len(r3Clean)))

		if idx < len(packages)-1 {
			fmt.Println(tui.RenderEmptyLine())
		}
	}

	fmt.Println(tui.RenderEmptyLine())
	fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))

	lUnsub := "  [unsub N] Berhenti berlangganan paket tertentu (contoh: unsub 1)"
	fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("unsub N", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Berhenti berlangganan paket tertentu (contoh: unsub 1)"), len(lUnsub)))

	lBack := "  [00] Kembali ke Menu Utama"
	fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

	fmt.Println(tui.RenderEmptyLine())
	fmt.Println(tui.RenderBoxBottom("Ketik opsi", tui.Width))
	fmt.Println()

	choice := tui.TuiInput("Pilihan paket")
	if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
		return
	}

	if strings.HasPrefix(choice, "unsub") {
		parts := strings.Fields(choice)
		if len(parts) == 2 {
			idx, err := strconv.Atoi(parts[1])
			if err == nil && idx >= 1 && idx <= len(packages) {
				target, _ := packages[idx-1].(map[string]any)
				optCode, _ := target["package_option_code"].(string)
				name, _ := target["name"].(string)
				if tui.TuiConfirm(fmt.Sprintf("Yakin ingin berhenti berlangganan paket '%s'?", name), false) {
					tui.RenderAlert("Mengirim permintaan berhenti berlangganan...", "info")
					_, err := client.Unsubscribe(idToken, activeUser.SubscriberID, optCode)
					if err != nil {
						tui.RenderAlert(fmt.Sprintf("Gagal berhenti berlangganan: %v", err), "error")
					} else {
						tui.RenderAlert("Berhasil berhenti berlangganan!", "success")
					}
					tui.TuiPause()
				}
			}
		}
	}
}

func GetPackagesByFamily(familyCode string, isEnterprise ...bool) {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	ent := false
	if len(isEnterprise) > 0 {
		ent = isEnterprise[0]
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	tui.ClearScreen()
	tui.RenderAlert(fmt.Sprintf("Mengambil data family '%s'...", familyCode), "info")

	familyData, err := client.GetFamily(idToken, familyCode, ent)
	if err != nil || familyData == nil {
		tui.RenderAlert("Gagal memuat data family dari server.", "error")
		tui.TuiPause()
		return
	}

	variants, _ := familyData["package_variants"].([]any)
	if len(variants) == 0 {
		tui.RenderAlert("Tidak ada varian paket ditemukan di family ini.", "warning")
		tui.TuiPause()
		return
	}

	type OptionEntry struct {
		VariantName string
		OptionName  string
		OptionCode  string
		Price       int
		Order       int
	}

	var allOptions []OptionEntry
	for _, v := range variants {
		vMap, _ := v.(map[string]any)
		if vMap == nil {
			continue
		}
		varName, _ := vMap["name"].(string)
		opts, _ := vMap["package_options"].([]any)
		for _, opt := range opts {
			optMap, _ := opt.(map[string]any)
			if optMap == nil {
				continue
			}
			optName, _ := optMap["name"].(string)
			optCode, _ := optMap["package_option_code"].(string)
			orderVal := 0
			switch ov := optMap["order"].(type) {
			case float64:
				orderVal = int(ov)
			case int:
				orderVal = ov
			}
			priceVal := 0
			switch pv := optMap["price"].(type) {
			case float64:
				priceVal = int(pv)
			case int:
				priceVal = pv
			}

			allOptions = append(allOptions, OptionEntry{
				VariantName: varName,
				OptionName:  optName,
				OptionCode:  optCode,
				Price:       priceVal,
				Order:       orderVal,
			})
		}
	}

	for {
		tui.ClearScreen()
		tui.RenderHeader("DAFTAR OPSI PAKET", "engsel ❯ Family", tui.Width)
		fmt.Println(tui.RenderHdrLine(fmt.Sprintf("OPSI PAKET (%d)", len(allOptions)), tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		for idx, item := range allOptions {
			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			fullTitle := fmt.Sprintf("%s - %s", item.VariantName, item.OptionName)
			priceStr := tui.FormatRupiah(item.Price)

			availW := (tui.Width - 2) - len(fmt.Sprintf("  [%2d] ", idx+1)) - len(priceStr) - 2
			titleClean := tui.TruncateStr(fullTitle, availW)

			leftP := fmt.Sprintf("  [%2d] %s", idx+1, titleClean)
			rem := (tui.Width - 2) - len(leftP) - len(priceStr) - 1
			if rem < 1 {
				rem = 1
			}

			row1 := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(titleClean) + tui.BgSpaces(rem) + tui.CAmb(priceStr) + tui.BgSpaces(1)
			fmt.Println(tui.PLine(row1, len(leftP)+rem+len(priceStr)+1))
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))
		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih nomor paket (1..N)", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilih paket (nomor)")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		}

		idx, err := strconv.Atoi(choice)
		if err == nil && idx >= 1 && idx <= len(allOptions) {
			sel := allOptions[idx-1]
			ShowPackageDetails(sel.OptionCode, ent, sel.Order)
		} else {
			tui.RenderAlert("Nomor paket tidak valid.", "warning")
			tui.TuiPause()
		}
	}
}
