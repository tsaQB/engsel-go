package menu

import (
	"fmt"
	"strconv"

	"github.com/tsaQB/engsel-go/internal/client/circle"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/crypto"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func ShowCircleCreation() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		return
	}

	tui.ClearScreen()
	tui.RenderHeader("BUAT CIRCLE BARU", "engsel ❯ Circle ❯ Buat", tui.Width)
	fmt.Println(tui.RenderEmptyLine())

	parentName := tui.TuiInput("Nama Anda (Parent/Ketua)")
	groupName := tui.TuiInput("Nama Circle / Komunitas")
	rawMember := tui.TuiInput("Nomor Anggota Pertama (contoh: 08123456789)")
	memberMSISDN := config.NormalizeMSISDN(rawMember)
	memberName := tui.TuiInput("Nama Anggota Pertama")

	if parentName == "" || groupName == "" || memberMSISDN == "" || memberName == "" {
		tui.RenderAlert("Semua data harus diisi lengkap untuk membuat Circle.", "error")
		tui.TuiPause()
		return
	}

	tui.RenderAlert("Membuat Circle baru di sistem MyXL...", "info")
	res, err := circle.CreateCircle(activeUser.Tokens, parentName, groupName, memberMSISDN, memberName)
	if err != nil || res == nil {
		tui.RenderAlert(fmt.Sprintf("Gagal membuat Circle: %v", err), "error")
	} else {
		status, _ := res["status"].(string)
		if status == "SUCCESS" {
			tui.RenderAlert(fmt.Sprintf("Circle '%s' berhasil dibuat!", groupName), "success")
		} else {
			tui.RenderAlert(fmt.Sprintf("Gagal membuat Circle: %v", res), "error")
		}
	}
	tui.TuiPause()
}

func ShowCircleInfo() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	for {
		tui.ClearScreen()
		tui.RenderAlert("Mengambil data Circle dari server...", "info")

		res, err := circle.GetGroupData(activeUser.Tokens)
		if err != nil || res == nil {
			tui.RenderAlert("Gagal mengambil data Circle.", "error")
			tui.TuiPause()
			return
		}

		dataMap, _ := res["data"].(map[string]any)
		groupID, _ := dataMap["group_id"].(string)
		groupName, _ := dataMap["group_name"].(string)

		if groupID == "" {
			tui.ClearScreen()
			tui.RenderHeader("CIRCLE GROUP", "engsel ❯ Circle", tui.Width)
			tui.RenderAlert("Anda belum bergabung atau memiliki Circle.", "info")
			if tui.TuiConfirm("Ingin membuat Circle baru sekarang?", true) {
				ShowCircleCreation()
			} else {
				return
			}
			continue
		}

		membersRes, _ := circle.GetGroupMembers(activeUser.Tokens, groupID)
		var membersList []any
		if membersRes != nil {
			if d, ok := membersRes["data"].(map[string]any); ok {
				if m, ok := d["members"].([]any); ok {
					membersList = m
				}
			}
		}

		parentMemberID := ""
		for _, m := range membersList {
			mMap, _ := m.(map[string]any)
			if mMap == nil {
				continue
			}
			role, _ := mMap["role"].(string)
			if role == "PARENT" || role == "OWNER" {
				parentMemberID, _ = mMap["member_id"].(string)
				break
			}
		}

		tui.ClearScreen()
		tui.RenderHeader("INFORMASI CIRCLE", "engsel ❯ Circle", tui.Width)

		fmt.Println(tui.RenderHdrLine("DETAIL CIRCLE", tui.Width))
		fmt.Println(tui.RenderKVLine("Nama Circle", groupName, tui.CWhi, tui.Width))
		fmt.Println(tui.RenderKVLine("Group ID", groupID, tui.CDim, tui.Width))
		fmt.Println(tui.RenderEmptyLine())

		fmt.Println(tui.RenderBoxDivider(fmt.Sprintf("ANGGOTA CIRCLE (%d)", len(membersList)), tui.Width))

		for idx, m := range membersList {
			mMap, _ := m.(map[string]any)
			if mMap == nil {
				continue
			}

			name, _ := mMap["name"].(string)
			encMSISDN, _ := mMap["msisdn"].(string)
			role, _ := mMap["role"].(string)
			decMSISDN := crypto.DecryptCircleMSISDN(encMSISDN)
			if decMSISDN == "" {
				decMSISDN = encMSISDN
			}

			kBadge := tui.Badge(fmt.Sprintf("%2d", idx+1), tui.CPri)
			rowP := fmt.Sprintf("  [%2d] %s (%s) [%s]", idx+1, name, tui.FormatMSISDN(decMSISDN), role)
			content := tui.BgSpaces(2) + kBadge + tui.BgSpaces(1) + tui.CWhi(name) + tui.BgSpaces(1) + tui.CMut("("+tui.FormatMSISDN(decMSISDN)+")") + tui.BgSpaces(1) + tui.CDim("["+role+"]")
			fmt.Println(tui.PLine(content, len(rowP)))
		}

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxDivider("OPSI CIRCLE", tui.Width))

		l1 := "  [ 1] Undang Anggota Baru"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Undang Anggota Baru"), len(l1)))

		l2 := "  [ 2] Hapus Anggota dari Circle"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("2", tui.CDan)+tui.BgSpaces(1)+tui.CWhi("Hapus Anggota dari Circle"), len(l2)))

		l3 := "  [ 3] Lihat Bonus & Hadiah Circle"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("3", tui.CSuc)+tui.BgSpaces(1)+tui.CWhi("Lihat Bonus & Hadiah Circle"), len(l3)))

		l4 := "  [ 4] Spending Tracker"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("4", tui.CAmb)+tui.BgSpaces(1)+tui.CWhi("Spending Tracker"), len(l4)))

		l5 := "  [ 5] Terima Undangan Circle"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("5", tui.CPur)+tui.BgSpaces(1)+tui.CWhi("Terima Undangan Circle"), len(l5)))

		lBack := "  [00] Kembali ke Menu Utama"
		fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

		fmt.Println(tui.RenderEmptyLine())
		fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
		fmt.Println()

		choice := tui.TuiInput("Pilihan")
		if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
			return
		} else if choice == "1" {
			rawM := tui.TuiInput("Nomor yang diundang (contoh: 08123456789)")
			msisdn := config.NormalizeMSISDN(rawM)
			name := tui.TuiInput("Nama Anggota")
			if msisdn != "" && name != "" {
				tui.RenderAlert("Mengirim undangan...", "info")
				res, err := circle.InviteCircleMember(activeUser.Tokens, msisdn, name, groupID, parentMemberID)
				if err != nil || res == nil {
					tui.RenderAlert(fmt.Sprintf("Gagal mengundang anggota: %v", err), "error")
				} else {
					tui.RenderAlert("Undangan berhasil dikirim!", "success")
				}
				tui.TuiPause()
			}
		} else if choice == "2" {
			numStr := tui.TuiInput("Nomor urut anggota yang ingin dihapus (1..N)")
			idx, err := strconv.Atoi(numStr)
			if err == nil && idx >= 1 && idx <= len(membersList) {
				target := membersList[idx-1].(map[string]any)
				targetMemberID, _ := target["member_id"].(string)
				if tui.TuiConfirm("Yakin ingin menghapus anggota ini?", false) {
					tui.RenderAlert("Menghapus anggota...", "info")
					res, err := circle.RemoveCircleMember(activeUser.Tokens, targetMemberID, groupID, parentMemberID, len(membersList) == 1)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal menghapus anggota: %v", err), "error")
					} else {
						tui.RenderAlert("Anggota berhasil dihapus!", "success")
					}
					tui.TuiPause()
				}
			}
		} else if choice == "3" {
			tui.RenderAlert("Mengambil data bonus Circle...", "info")
			bonusRes, err := circle.GetBonusData(activeUser.Tokens, activeUser.SubscriberID, groupID)
			if err != nil || bonusRes == nil {
				tui.RenderAlert("Gagal mengambil bonus Circle.", "error")
			} else {
				tui.ClearScreen()
				tui.RenderHeader("BONUS CIRCLE", "engsel ❯ Circle ❯ Bonus", tui.Width)
				fmt.Printf("  Status: %v\n", bonusRes["status"])
				if d, ok := bonusRes["data"]; ok {
					fmt.Printf("  Data: %v\n", d)
				}
			}
			tui.TuiPause()
		} else if choice == "4" {
			tui.RenderAlert("Mengambil data spending tracker...", "info")
			spRes, err := circle.SpendingTracker(activeUser.Tokens, groupID)
			if err != nil || spRes == nil {
				tui.RenderAlert("Gagal mengambil spending tracker.", "error")
			} else {
				tui.ClearScreen()
				tui.RenderHeader("SPENDING TRACKER", "engsel ❯ Circle ❯ Tracker", tui.Width)
				if d, ok := spRes["data"]; ok {
					fmt.Printf("  Data: %v\n", d)
				}
			}
			tui.TuiPause()
		} else if choice == "5" {
			gID := tui.TuiInput("Masukkan Group ID yang mengundang Anda")
			if gID != "" {
				tui.RenderAlert("Menerima undangan...", "info")
				res, err := circle.AcceptCircleInvitation(activeUser.Tokens, gID)
				if err != nil || res == nil {
					tui.RenderAlert(fmt.Sprintf("Gagal menerima undangan: %v", err), "error")
				} else {
					tui.RenderAlert("Berhasil bergabung dengan Circle!", "success")
				}
				tui.TuiPause()
			}
		}
	}
}
