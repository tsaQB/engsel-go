package menu

import (
	"fmt"
	"strconv"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/client/sharing"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowBalanceAllotmentMenu() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu untuk transfer pulsa.", "error")
		tui.TuiPause()
		return
	}

	tui.ClearScreen()
	tui.RenderHeader("TRANSFER PULSA", "engsel ❯ Sharing", tui.Width)

	idToken, _ := activeUser.Tokens["id_token"].(string)
	accessToken, _ := activeUser.Tokens["access_token"].(string)

	balanceRes, _ := client.GetBalance(idToken)
	var balanceRemaining int64
	if balanceRes != nil {
		switch v := balanceRes["remaining"].(type) {
		case float64:
			balanceRemaining = int64(v)
		case int64:
			balanceRemaining = v
		}
	}

	senderFmt := tui.FormatMSISDN(activeUser.Number)
	balFmt := tui.FormatRupiah(balanceRemaining)

	fmt.Println(tui.RenderHdrLine("INFORMASI PENGIRIM", tui.Width))
	fmt.Println(tui.RenderKVLine("Nomor Anda", senderFmt, tui.CWhi, tui.Width))
	fmt.Println(tui.RenderKVLine("Sisa Pulsa", balFmt, tui.CAmb, tui.Width))
	fmt.Println(tui.RenderEmptyLine())

	infoLine := "  Catatan: Pastikan Anda sudah mengatur PIN di MyXL."
	fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.CMut(infoLine[2:]), len(infoLine)))
	fmt.Println(tui.RenderEmptyLine())
	fmt.Println(tui.RenderBoxBottom("", tui.Width))
	fmt.Println()

	pin := tui.TuiInput("PIN Transaksi (6 digit)")
	if pin == "00" || pin == "q" || pin == "" {
		return
	}
	if len(pin) != 6 {
		tui.RenderAlert("Format PIN tidak valid! Harus berupa 6 digit angka.", "error")
		tui.TuiPause()
		return
	}

	tui.RenderAlert("Memvalidasi PIN & otentikasi transaksi...", "info")
	stageToken, err := client.GetAuthCode(accessToken, pin, fmt.Sprintf("%d", activeUser.Number))
	if err != nil || stageToken == "" {
		tui.RenderAlert("Gagal mendapatkan kode otentikasi. Pastikan PIN Anda benar.", "error")
		tui.TuiPause()
		return
	}

	receiverInput := tui.TuiInput("Nomor Penerima (contoh: 08123456789)")
	if receiverInput == "" || receiverInput == "00" || receiverInput == "q" {
		return
	}
	receiverMsisdn := config.NormalizeMSISDN(receiverInput)

	amountStr := tui.TuiInput("Nominal Transfer (contoh: 5000)")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		tui.RenderAlert("Nominal transfer tidak valid! Masukkan angka lebih dari 0.", "error")
		tui.TuiPause()
		return
	}

	tui.ClearScreen()
	tui.RenderHeader("KONFIRMASI TRANSFER", "engsel ❯ Konfirmasi", tui.Width)
	fmt.Println(tui.RenderHdrLine("RINCIAN TRANSAKSI", tui.Width))
	fmt.Println(tui.RenderKVLine("Penerima", tui.FormatMSISDN(receiverMsisdn), tui.CWhi, tui.Width))
	fmt.Println(tui.RenderKVLine("Nominal", tui.FormatRupiah(amount), tui.CAmb, tui.Width))
	fmt.Println(tui.RenderKVLine("Sisa Nanti", tui.FormatRupiah(balanceRemaining-int64(amount)), tui.CMut, tui.Width))
	fmt.Println(tui.RenderEmptyLine())
	fmt.Println(tui.RenderBoxBottom("", tui.Width))
	fmt.Println()

	if !tui.TuiConfirm("Lanjutkan proses transfer pulsa ini?", false) {
		tui.RenderAlert("Transfer pulsa dibatalkan oleh pengguna.", "info")
		tui.TuiPause()
		return
	}

	tui.RenderAlert("Mengirim permintaan transfer pulsa ke server...", "info")
	res, err := sharing.BalanceAllotment(activeUser.Tokens, stageToken, receiverMsisdn, amount)
	if err != nil || res == nil {
		tui.RenderAlert(fmt.Sprintf("Transfer pulsa gagal diproses: %v", err), "error")
	} else {
		tui.RenderAlert(fmt.Sprintf("Transfer pulsa %s ke %s berhasil dikirim!", tui.FormatRupiah(amount), tui.FormatMSISDN(receiverMsisdn)), "success")
	}
	tui.TuiPause()
}
