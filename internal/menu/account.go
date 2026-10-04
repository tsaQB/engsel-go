package menu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func LoginPrompt() (int64, string) {
	tui.ClearScreen()
	tui.RenderHeader("LOGIN MYXL", "engsel ❯ Login", tui.Width)
	fmt.Println(tui.RenderEmptyLine())

	info := "  Masukkan nomor XL Anda (contoh: 081234567890)"
	fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(info[2:]), len(info)))
	fmt.Println(tui.RenderEmptyLine())
	fmt.Println(tui.RenderBoxBottom("", tui.Width))
	fmt.Println()

	rawNumber := tui.TuiInput("Nomor XL")
	if rawNumber == "" || rawNumber == "00" || rawNumber == "0" || rawNumber == "q" {
		return 0, ""
	}

	phone := config.NormalizeMSISDN(rawNumber)
	if !strings.HasPrefix(phone, "628") || len(phone) < 10 || len(phone) > 14 {
		tui.RenderAlert("Nomor tidak valid! Pastikan diawali 08/628 dan panjang sesuai.", "error")
		tui.TuiPause()
		return 0, ""
	}

	subscriberID, err := client.GetOTP(phone)
	if err != nil || subscriberID == "" {
		tui.RenderAlert(fmt.Sprintf("Gagal meminta OTP: %v", err), "error")
		tui.TuiPause()
		return 0, ""
	}

	tui.RenderAlert(fmt.Sprintf("OTP berhasil dikirim via SMS ke %s", tui.FormatMSISDN(phone)), "success")

	tryCount := 5
	for tryCount > 0 {
		fmt.Println()
		lvl := "info"
		if tryCount <= 2 {
			lvl = "warning"
		}
		tui.RenderAlert(fmt.Sprintf("Sisa kesempatan memasukkan OTP: %d kali", tryCount), lvl)

		otp := tui.TuiInput("Kode OTP (6 digit)")
		if otp == "00" || otp == "q" {
			tui.RenderAlert("Proses login dibatalkan oleh pengguna.", "warning")
			tui.TuiPause()
			return 0, ""
		}

		if len(otp) != 6 {
			tui.RenderAlert("OTP tidak valid! Harus berupa 6 digit angka.", "error")
			continue
		}

		tokens, err := client.SubmitOTP("SMS", phone, otp, "")
		if err != nil || tokens == nil {
			tui.RenderAlert("Kode OTP salah atau telah kedaluwarsa. Silakan coba lagi.", "error")
			tryCount--
			continue
		}

		refreshToken, _ := tokens["refresh_token"].(string)
		num, _ := strconv.ParseInt(phone, 10, 64)
		tui.RenderAlert(fmt.Sprintf("Login berhasil untuk nomor %s!", tui.FormatMSISDN(phone)), "success")
		tui.TuiPause()
		return num, refreshToken
	}

	tui.RenderAlert("Gagal login setelah 5 percobaan. Silakan coba lagi nanti.", "error")
	tui.TuiPause()
	return 0, ""
}

func ShowAccountMenu() int64 {
	addUser := false

	for {
		tui.ClearScreen()
		activeUser := service.AuthInstance.GetActiveUser()
		users := service.AuthInstance.RefreshTokens

		if activeUser == nil || addUser {
			num, refreshToken := LoginPrompt()
			if refreshToken == "" {
				tui.RenderAlert("Gagal menambah akun baru.", "warning")
				tui.TuiPause()
				addUser = false
				continue
			}

			service.AuthInstance.AddRefreshToken(num, refreshToken)
			addUser = false
			continue
		}

		tui.RenderHeader("MANAJEMEN AKUN", "engsel ❯ Akun", tui.Width)
		fmt.Println(tui.RenderHdrLine("DAFTAR AKUN TERSIMPAN", tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		if len(users) == 0 {
			emptyStr := "  Belum ada akun XL yang tersimpan."
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(emptyStr[2:]), len(emptyStr)))
		} else {
			for idx, u := range users {
				isActive := activeUser != nil && u.Number == activeUser.Number
				numFmt := tui.FormatMSISDN(u.Number)
				subLabel := "Prabayar"
				if u.SubscriptionType == "POSTPAID" {
					subLabel = "Pascabayar"
				}

				badgeCol := tui.CPri
				numCol := tui.CMut
				if isActive {
					badgeCol = tui.CWhi
					numCol = tui.CWhi
				}

				statusCol := tui.BgSpaces(7)
				plainRow := fmt.Sprintf("  [%2d] %-14s [%-10s]        ", idx+1, numFmt, subLabel)
				if isActive {
					statusCol = tui.CSuc("✔ AKTIF")
					plainRow = fmt.Sprintf("  [%2d] %-14s [%-10s] ✔ AKTIF", idx+1, numFmt, subLabel)
				}

				content := tui.BgSpaces(2) +
					tui.Badge(fmt.Sprintf("%d", idx+1), badgeCol) +
					tui.BgSpaces(1) +
					numCol(fmt.Sprintf("%-14s", numFmt)) +
					tui.BgSpaces(1) +
					tui.CDim(fmt.Sprintf("[%-10s]", subLabel)) +
					tui.BgSpaces(1) +
					statusCol

				fmt.Println(tui.PLine(content, len(plainRow)))
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("AKSI & PERINTAH", tui.Width))

		lAdd := "  [ 0] Tambah Akun Baru (Login)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("0", tui.CSuc)+tui.BgSpaces(1)+tui.CWhi("Tambah Akun Baru (Login)"), len(lAdd)))

		lSwitch := "  [1..N] Ketik nomor urut untuk ganti akun aktif"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1..N", tui.CPri)+tui.BgSpaces(1)+tui.CMut("Ketik nomor urut untuk ganti akun aktif"), len(lSwitch)))

		lDel := "  [del N] Hapus akun tertentu (contoh: del 2)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("del N", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Hapus akun tertentu (contoh: del 2)"), len(lDel)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Ketik nomor/perintah", tui.Width))
		fmt.Println()

		inputStr := tui.TuiInput("Pilihan")
		if inputStr == "00" || inputStr == "q" || inputStr == "exit" {
			if activeUser != nil {
				return activeUser.Number
			}
			return 0
		} else if inputStr == "0" {
			addUser = true
			continue
		} else if strings.HasPrefix(inputStr, "del") {
			parts := strings.Fields(inputStr)
			if len(parts) == 2 {
				idx, err := strconv.Atoi(parts[1])
				if err == nil && idx >= 1 && idx <= len(users) {
					target := users[idx-1]
					if activeUser != nil && target.Number == activeUser.Number {
						tui.RenderAlert("Tidak dapat menghapus akun yang sedang aktif! Beralih ke akun lain dulu.", "error")
						tui.TuiPause()
						continue
					}
					if tui.TuiConfirm(fmt.Sprintf("Yakin ingin menghapus akun %s?", tui.FormatMSISDN(target.Number)), false) {
						service.AuthInstance.RemoveRefreshToken(target.Number)
						tui.RenderAlert("Akun berhasil dihapus.", "success")
					} else {
						tui.RenderAlert("Penghapusan dibatalkan.", "info")
					}
					tui.TuiPause()
					continue
				}
			}
			tui.RenderAlert("Format salah! Gunakan: del <nomor urut> (contoh: del 2)", "warning")
			tui.TuiPause()
		} else {
			idx, err := strconv.Atoi(inputStr)
			if err == nil && idx >= 1 && idx <= len(users) {
				sel := users[idx-1]
				service.AuthInstance.SetActiveUser(sel.Number)
				tui.RenderAlert(fmt.Sprintf("Beralih ke akun %s...", tui.FormatMSISDN(sel.Number)), "success")
				return sel.Number
			} else {
				tui.RenderAlert("Pilihan tidak valid. Silakan coba lagi.", "error")
				tui.TuiPause()
			}
		}
	}
}
