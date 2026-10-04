package menu

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tsaQB/engsel-go/internal/client/famplan"

	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowFamilyInfo() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil data Akrab Organizer dari server...", "info")

		res, err := famplan.GetFamilyData(activeUser.Tokens)
		if err != nil || res == nil {
			tui.RenderAlert("Gagal mengambil data keluarga dari server.", "error")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		memberInfo, _ := dataMap["member_info"].(map[string]any)
		planType, _ := memberInfo["plan_type"].(string)

		if planType == "" {
			tui.RenderAlert("Akun Anda bukan merupakan Pengelola (Organizer) Paket Akrab.", "warning")
			tui.TuiPause()
			return
		}

		parentMsisdn, _ := memberInfo["parent_msisdn"].(string)
		members, _ := memberInfo["members"].([]any)

		var totalQuota int64
		var remainingQuota int64
		switch v := memberInfo["total_quota"].(type) {
		case float64:
			totalQuota = int64(v)
		case int64:
			totalQuota = v
		}
		switch v := memberInfo["remaining_quota"].(type) {
		case float64:
			remainingQuota = int64(v)
		case int64:
			remainingQuota = v
		}

		var endDateTs int64
		switch v := memberInfo["end_date"].(type) {
		case float64:
			endDateTs = int64(v)
		case int64:
			endDateTs = v
		}
		endDate := "-"
		if endDateTs > 0 {
			endDate = time.Unix(endDateTs, 0).Format("02 Jan 2006")
		}

		tui.ClearScreen()
		tui.RenderHeader("AKRAB ORGANIZER", "engsel ❯ Akrab", tui.Width)

		fmt.Println(tui.RenderHdrLine("INFORMASI GRUP AKRAB", tui.Width))
		fmt.Println(tui.RenderKVLine("Paket Akrab", planType, tui.CWhi, tui.Width))
		fmt.Println(tui.RenderKV2Col("Pengelola", tui.FormatMSISDN(parentMsisdn), tui.CWhi, "Berlaku s/d", endDate, tui.CDim, 28, tui.Width))
		fmt.Println(tui.RenderKV2Col("Kuota Sisa", tui.FormatQuotaByte(remainingQuota), tui.CAmb, "Total Kuota", tui.FormatQuotaByte(totalQuota), tui.CMut, 28, tui.Width))

		// Quota Bar
		pct := 0
		if totalQuota > 0 {
			pct = int((remainingQuota * 100) / totalQuota)
		}
		barStr := tui.RenderProgress(remainingQuota, totalQuota, 20)
		rowBar := tui.BgSpaces(2) + tui.CMut("Sisa Kuota: ") + barStr + tui.CWhi(fmt.Sprintf(" %d%%", pct))
		fmt.Println(tui.PLine(rowBar, len(fmt.Sprintf("  Sisa Kuota: [....................] %d%%", pct))))
		fmt.Println(tui.RenderEmptyLine())

		// Members
		fmt.Println(tui.RenderBoxDivider(fmt.Sprintf("SLOT ANGGOTA (%d)", len(members)), tui.Width))

		for idx, m := range members {
			mMap, _ := m.(map[string]any)
			if mMap == nil {
				continue
			}

			msisdn, _ := mMap["msisdn"].(string)
			alias, _ := mMap["alias"].(string)
			isEmpty := msisdn == ""

			usage, _ := mMap["usage"].(map[string]any)
			var quotaAlloc int64
			var quotaUsed int64
			if usage != nil {
				switch v := usage["quota_allocated"].(type) {
				case float64:
					quotaAlloc = int64(v)
				case int64:
					quotaAlloc = v
				}
				switch v := usage["quota_used"].(type) {
				case float64:
					quotaUsed = int64(v)
				case int64:
					quotaUsed = v
				}
			}

			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			if isEmpty {
				rowP := fmt.Sprintf("  [%2d] [SLOT KOSONG]", idx+1)
				content := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CDim("[SLOT KOSONG]")
				fmt.Println(tui.PLine(content, len(rowP)))
			} else {
				row1P := fmt.Sprintf("  [%2d] %s (%s)", idx+1, tui.FormatMSISDN(msisdn), alias)
				content1 := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(tui.FormatMSISDN(msisdn)) + tui.BgSpaces(1) + tui.CMut("("+alias+")")
				fmt.Println(tui.PLine(content1, len(row1P)))

				row2P := fmt.Sprintf("      Terpakai: %s / Limit: %s", tui.FormatQuotaByte(quotaUsed), tui.FormatQuotaByte(quotaAlloc))
				content2 := tui.BgSpaces(6) + tui.CMut("Terpakai: ") + tui.CAmb(tui.FormatQuotaByte(quotaUsed)) + tui.CMut(" / Limit: ") + tui.CWhi(tui.FormatQuotaByte(quotaAlloc))
				fmt.Println(tui.PLine(content2, len(row2P)))
			}
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI AKRAB", tui.Width))

		l1 := "  [ 1] Ubah Anggota (Ganti Nomor)"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Ubah Anggota (Ganti Nomor)"), len(l1)))

		l2 := "  [ 2] Hapus Anggota dari Slot"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("2", tui.CDan)+tui.BgSpaces(1)+tui.CWhi("Hapus Anggota dari Slot"), len(l2)))

		l3 := "  [ 3] Atur Batas (Limit) Kuota Anggota"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("3", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Atur Batas (Limit) Kuota Anggota"), len(l3)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		} else if choice == "1" {
			slotStr := tui.TuiInput("Pilih nomor slot yang ingin diubah (1..N)")
			slotIdx, err := strconv.Atoi(slotStr)
			if err == nil && slotIdx >= 1 && slotIdx <= len(members) {
				m := members[slotIdx-1].(map[string]any)
				newNumRaw := tui.TuiInput("Nomor baru (contoh: 08123456789)")
				newMsisdn := config.NormalizeMSISDN(newNumRaw)
				alias := tui.TuiInput("Nama alias anggota baru", "Anggota")
				famMemberID, _ := m["family_member_id"].(string)

				slotID := slotIdx
				switch v := m["slot_id"].(type) {
				case float64:
					slotID = int(v)
				case int:
					slotID = v
				}

				tui.RenderAlert("Memproses pergantian anggota...", "info")
				res, err := famplan.ChangeMember(activeUser.Tokens, "Parent", alias, slotID, famMemberID, newMsisdn)
				if err != nil || res == nil {
					tui.RenderAlert(fmt.Sprintf("Gagal mengubah anggota: %v", err), "error")
				} else {
					tui.RenderAlert("Anggota berhasil diperbarui!", "success")
				}
				tui.TuiPause()
			}
		} else if choice == "2" {
			slotStr := tui.TuiInput("Pilih nomor slot yang ingin dihapus (1..N)")
			slotIdx, err := strconv.Atoi(slotStr)
			if err == nil && slotIdx >= 1 && slotIdx <= len(members) {
				m := members[slotIdx-1].(map[string]any)
				famMemberID, _ := m["family_member_id"].(string)
				slotID := slotIdx
				switch v := m["slot_id"].(type) {
				case float64:
					slotID = int(v)
				case int:
					slotID = v
				}

				if tui.TuiConfirm("Yakin ingin menghapus anggota ini?", false) {
					tui.RenderAlert("Menghapus anggota...", "info")
					res, err := famplan.RemoveMember(activeUser.Tokens, "Parent", slotID, famMemberID)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal menghapus anggota: %v", err), "error")
					} else {
						tui.RenderAlert("Anggota berhasil dihapus dari grup!", "success")
					}
					tui.TuiPause()
				}
			}
		} else if choice == "3" {
			slotStr := tui.TuiInput("Pilih nomor slot (1..N)")
			slotIdx, err := strconv.Atoi(slotStr)
			if err == nil && slotIdx >= 1 && slotIdx <= len(members) {
				m := members[slotIdx-1].(map[string]any)
				limitGBStr := tui.TuiInput("Limit kuota dalam GB (contoh: 5)")
				limitGB, err := strconv.ParseFloat(limitGBStr, 64)
				if err == nil && limitGB >= 0 {
					limitBytes := int64(limitGB * 1024 * 1024 * 1024)
					famMemberID, _ := m["family_member_id"].(string)
					alias, _ := m["alias"].(string)
					slotID := slotIdx
					switch v := m["slot_id"].(type) {
					case float64:
						slotID = int(v)
					case int:
						slotID = v
					}

					tui.RenderAlert("Memperbarui limit kuota anggota...", "info")
					res, err := famplan.SetQuotaLimit(activeUser.Tokens, "Parent", alias, slotID, famMemberID, limitBytes)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal mengatur kuota: %v", err), "error")
					} else {
						tui.RenderAlert("Limit kuota berhasil diperbarui!", "success")
					}
					tui.TuiPause()
				}
			}
		}
	}
}
