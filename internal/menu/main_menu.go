package menu

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/client/famplan"
	"github.com/tsaQB/engsel-go/internal/client/registration"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

type UserProfile struct {
	Number           int64
	SubscriberID     string
	SubscriptionType string
	Balance          int64
	BalanceExpiredAt int64
	Tier             int
	Points           int
	TierName         string
}

func RenderMainMenuDisplay(p UserProfile) {
	tui.RenderBanner("CLI Automation Suite for MyXL", tui.Width)

	msisdnFormatted := tui.FormatMSISDN(p.Number)
	balanceFormatted := tui.FormatRupiah(p.Balance)
	subLabel := "Prabayar"
	if p.SubscriptionType == "POSTPAID" {
		subLabel = "Pascabayar"
	}

	expiredStr := "-"
	if p.BalanceExpiredAt > 0 {
		expiredStr = time.Unix(p.BalanceExpiredAt, 0).Format("2006-01-02")
	}

	tierInfo := fmt.Sprintf("%s (%d)", p.TierName, p.Points)

	// Top Border Panel Box
	fmt.Println(tui.Border("┌" + strings.Repeat("─", tui.Width-2) + "┐"))

	// Account Profile Card
	fmt.Println(tui.RenderCardRow("Nomor", msisdnFormatted, tui.CWhi, "Pulsa", balanceFormatted, tui.CAmb, 28, tui.Width))
	fmt.Println(tui.RenderCardRow("Tipe", subLabel, tui.CMut, "Tier", tierInfo, tui.CWhi, 28, tui.Width))
	fmt.Println(tui.RenderCardRow("Status", "Terverifikasi", tui.CSuc, "Expired", expiredStr, tui.CDim, 28, tui.Width))

	// Divider
	fmt.Println(tui.Border("├" + strings.Repeat("─", tui.Width-2) + "┤"))

	// Section 1: Paket & Pembelian (1 - 10)
	fmt.Println(tui.RenderHdrLine("PAKET & PEMBELIAN", tui.Width))
	fmt.Println(tui.RenderMenuRow("1", "Lihat Paket Saya", "6", "Beli Paket HOT", 28, true, tui.Width))
	fmt.Println(tui.RenderMenuRow("2", "Beli Paket HOT-2", "7", "Beli Option Code", 28, true, tui.Width))
	fmt.Println(tui.RenderMenuRow("3", "Beli Family Code", "8", "Beli Borongan (Loop)", 28, true, tui.Width))
	fmt.Println(tui.RenderMenuRow("4", "Store Segments", "9", "Store Family List", 28, true, tui.Width))
	fmt.Println(tui.RenderMenuRow("5", "Store Packages", "10", "Redeemables", 28, true, tui.Width))
	fmt.Println(tui.PLine("", 0, tui.Width))

	// Section 2: Kuota Bersama & Sharing (11 - 14)
	fmt.Println(tui.RenderHdrLine("KUOTA BERSAMA & SHARING", tui.Width))
	fmt.Println(tui.RenderMenuRow("11", "Family Hub (Akrab)", "13", "Circle Group", 28, true, tui.Width))
	fmt.Println(tui.RenderMenuRow("12", "Akrab Organizer", "14", "Transfer Pulsa", 28, true, tui.Width))
	fmt.Println(tui.PLine("", 0, tui.Width))

	// Section 3: Akun & Utilitas (S, H, B, N, V, R)
	fmt.Println(tui.RenderHdrLine("AKUN & SISTEM", tui.Width))
	fmt.Println(tui.RenderMenuRow("S", "Ganti Akun / Switch", "H", "Riwayat Transaksi", 28, false, tui.Width))
	fmt.Println(tui.RenderMenuRow("B", "Bookmark Paket", "N", "Pusat Notifikasi", 28, false, tui.Width))
	fmt.Println(tui.RenderMenuRow("V", "Validasi Nomor", "R", "Registrasi Dukcapil", 28, false, tui.Width))
	fmt.Println(tui.PLine("", 0, tui.Width))

	// Exit Option
	exitC := tui.BgSpaces(2) + tui.CBor("[") + tui.CDan("99") + tui.CBor("]") + tui.BgSpaces(1) + tui.CDan("Tutup Aplikasi") + tui.BgSpaces(1) + tui.CDim("(atau tekan 0)")
	plainExit := len("  [99] Tutup Aplikasi (atau tekan 0)")
	fmt.Println(tui.PLine(exitC, plainExit, tui.Width))

	// Bottom Border Panel Box
	fmt.Println(tui.Border("└" + strings.Repeat("─", tui.Width-2) + "┘"))
	fmt.Println()
}

func RunMainMenu() {
	for {
		activeUser := service.AuthInstance.GetActiveUser()
		if activeUser == nil {
			selectedNum := ShowAccountMenu()
			if selectedNum != 0 {
				service.AuthInstance.SetActiveUser(selectedNum)
			}
			continue
		}

		idToken, _ := activeUser.Tokens["id_token"].(string)
		balanceRes, _ := client.GetBalance(idToken)

		var balanceRemaining int64
		var balanceExpiredAt int64
		if balanceRes != nil {
			switch v := balanceRes["remaining"].(type) {
			case float64:
				balanceRemaining = int64(v)
			case int64:
				balanceRemaining = v
			}
			switch v := balanceRes["expired_at"].(type) {
			case float64:
				balanceExpiredAt = int64(v)
			case int64:
				balanceExpiredAt = v
			}
		}

		tier := 0
		points := 0
		tierName := "Classic"

		if activeUser.SubscriptionType == "PREPAID" {
			tierRes, _ := client.GetTieringInfo(idToken)
			if tierRes != nil {
				dataMap, _ := tierRes["data"].(map[string]any)
				if dataMap == nil {
					dataMap = tierRes
				}
				switch v := dataMap["tier"].(type) {
				case float64:
					tier = int(v)
				case int:
					tier = v
				}
				switch v := dataMap["current_point"].(type) {
				case float64:
					points = int(v)
				case int:
					points = v
				}
				tierNames := map[int]string{0: "Classic", 1: "Silver", 2: "Gold", 3: "Platinum", 4: "Diamond"}
				if name, ok := tierNames[tier]; ok {
					tierName = name
				} else {
					tierName = fmt.Sprintf("Tier %d", tier)
				}
			}
		}

		profile := UserProfile{
			Number:           activeUser.Number,
			SubscriberID:     activeUser.SubscriberID,
			SubscriptionType: activeUser.SubscriptionType,
			Balance:          balanceRemaining,
			BalanceExpiredAt: balanceExpiredAt,
			Tier:             tier,
			Points:           points,
			TierName:         tierName,
		}

		tui.ClearScreen()
		RenderMainMenuDisplay(profile)

		choice := tui.TuiInput("Pilihan menu")
		choiceClean := strings.ToLower(strings.TrimSpace(choice))
		choiceKey := choiceClean
		if choiceClean != "0" && choiceClean != "00" {
			choiceKey = strings.TrimLeft(choiceClean, "0")
		}

		switch choiceKey {
		case "t":
			tui.TuiPause()
		case "1", "my", "pkg":
			FetchMyPackages()
		case "2", "hot2":
			ShowHotMenu2()
		case "3":
			familyCode := tui.TuiInput("Family Code (atau '00' untuk batal)")
			if familyCode != "" && familyCode != "00" && familyCode != "q" {
				GetPackagesByFamily(familyCode)
			}
		case "4":
			isEnt := tui.TuiConfirm("Akses katalog segmen enterprise?", false)
			ShowStoreSegmentsMenu(isEnt)
		case "5":
			isEnt := tui.TuiConfirm("Akses katalog paket enterprise?", false)
			ShowStorePackagesMenu(profile.SubscriptionType, isEnt)
		case "6", "hot":
			ShowHotMenu()
		case "7", "opt":
			optCode := tui.TuiInput("Option Code paket (atau '00' untuk batal)")
			if optCode != "" && optCode != "00" && optCode != "q" {
				ShowPackageDetails(optCode, false)
			}
		case "8", "loop":
			familyCode := tui.TuiInput("Family Code (atau '00' untuk batal)")
			if familyCode == "" || familyCode == "00" || familyCode == "q" {
				continue
			}
			startStr := tui.TuiInput("Mulai dari nomor opsi (default 1)", "1")
			startOpt, _ := strconv.Atoi(startStr)
			if startOpt <= 0 {
				startOpt = 1
			}
			useDecoy := tui.TuiConfirm("Gunakan paket decoy untuk bypass?", false)
			pauseOnSuccess := tui.TuiConfirm("Jeda (pause) setiap kali pembelian sukses?", true)
			delayStr := tui.TuiInput("Jeda antar pembelian dalam detik (0 untuk tanpa jeda)", "0")
			delaySec, _ := strconv.Atoi(delayStr)
			PurchaseByFamily(familyCode, useDecoy, pauseOnSuccess, delaySec, startOpt)
		case "9":
			isEnt := tui.TuiConfirm("Akses katalog enterprise?", false)
			ShowFamilyListMenu(profile.SubscriptionType, isEnt)
		case "10":
			isEnt := tui.TuiConfirm("Akses katalog enterprise?", false)
			ShowRedeemablesMenu(isEnt)
		case "11", "fh":
			FamilyHubMenu()
		case "12", "akrab":
			ShowFamilyInfo()
		case "13", "circle":
			ShowCircleInfo()
		case "14", "ba":
			ShowBalanceAllotmentMenu()
		case "s", "login", "acc", "15", "a":
			sel := ShowAccountMenu()
			if sel != 0 {
				service.AuthInstance.SetActiveUser(sel)
			}
		case "h", "history", "16":
			ShowTransactionHistory()
		case "b", "bm", "00", "17":
			ShowBookmarkMenu()
		case "n", "18":
			ShowNotificationMenu()
		case "v", "19":
			rawM := tui.TuiInput("Nomor MSISDN yang ingin divalidasi (contoh: 08123456789)")
			msisdn := config.NormalizeMSISDN(rawM)
			tui.RenderAlert(fmt.Sprintf("Memvalidasi nomor %s...", tui.FormatMSISDN(msisdn)), "info")
			res, err := famplan.ValidateMSISDN(activeUser.Tokens, msisdn)
			if err != nil || res == nil {
				tui.RenderAlert("Gagal memvalidasi nomor.", "error")
			} else {
				status, _ := res["status"].(string)
				if strings.ToLower(status) == "success" {
					role := "NO_ROLE"
					if d, ok := res["data"].(map[string]any); ok {
						if r, ok := d["family_plan_role"].(string); ok {
							role = r
						}
					}
					tui.RenderAlert(fmt.Sprintf("Hasil: %s | Role: %s | Status: VALID", tui.FormatMSISDN(msisdn), role), "success")
				} else {
					tui.RenderAlert(fmt.Sprintf("Validasi gagal: %v", res), "error")
				}
			}
			tui.TuiPause()
		case "r", "20":
			rawM := tui.TuiInput("Nomor MSISDN yang didaftarkan (contoh: 08123456789)")
			msisdn := config.NormalizeMSISDN(rawM)
			nik := tui.TuiInput("Nomor Induk Kependudukan (NIK 16 digit)")
			kk := tui.TuiInput("Nomor Kartu Keluarga (KK 16 digit)")

			tui.RenderAlert("Mengirim data registrasi Dukcapil...", "info")
			res, err := registration.Dukcapil(msisdn, kk, nik)
			if err != nil || res == nil {
				tui.RenderAlert("Gagal mengirim registrasi Dukcapil.", "error")
			} else {
				status, _ := res["status"].(string)
				if strings.ToLower(status) == "success" {
					tui.RenderAlert("Registrasi nomor berhasil diverifikasi oleh Dukcapil!", "success")
				} else {
					errCode, _ := res["code"].(string)
					rawMsg, _ := res["message"].(string)
					errMsg := fmt.Sprintf("Kode %s: %s", errCode, rawMsg)
					if errCode == "164" {
						errMsg = "Error 164: Kartu SIM belum terpasang di HP / belum latch sinyal BTS, atau NIK melebihi kuota 3 nomor."
					} else if errCode == "213" {
						errMsg = "Error 213: Data NIK atau Nomor KK tidak ditemukan / tidak cocok di Dukcapil."
					}
					tui.RenderAlert(fmt.Sprintf("Registrasi gagal: %s", errMsg), "error")
				}
			}
			tui.TuiPause()
		case "sentry":
			service.EnterSentryMode()
		case "99", "0", "q", "exit":
			tui.ClearScreen()
			tui.RenderAlert("Terima kasih telah menggunakan Engsel CLI TUI. Sampai jumpa!", "info")
			os.Exit(0)
		default:
			tui.RenderAlert("Pilihan menu tidak valid. Silakan pilih nomor atau kode yang tertera.", "warning")
			tui.TuiPause()
		}
	}
}
