package menu

import (
	"fmt"
	"strconv"

	"github.com/tsaQB/engsel-go/internal/client/familyhub"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/crypto"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func FamilyHubMenu() {
	activeUser := service.AuthInstance.GetActiveUser()
	if activeUser == nil {
		tui.RenderAlert("Silakan login terlebih dahulu.", "error")
		tui.TuiPause()
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)
	accessToken, _ := activeUser.Tokens["access_token"].(string)

	for {
		tui.ClearScreen()
		tui.RenderAlert("Memeriksa status Family Hub dari server...", "info")

		statusRes, err := familyhub.GetFamilyHubStatus(idToken)
		if err != nil || statusRes == nil {
			tui.RenderAlert("Gagal memeriksa status Family Hub.", "error")
			tui.TuiPause()
			return
		}

		dataMap, _ := statusRes["data"].(map[string]any)
		groupStatus, _ := dataMap["status"].(string)
		groupID, _ := dataMap["group_id"].(string)
		groupName, _ := dataMap["group_name"].(string)

		tui.ClearScreen()
		tui.RenderHeader("FAMILY HUB (AKRAB)", "engsel ❯ Family Hub", tui.Width)

		fmt.Println(tui.RenderHdrLine("STATUS FAMILY HUB", tui.Width))
		fmt.Println(tui.RenderKVLine("Status", groupStatus, tui.CWhi, tui.Width))
		if groupID != "" {
			fmt.Println(tui.RenderKVLine("Nama Grup", groupName, tui.CWhi, tui.Width))
			fmt.Println(tui.RenderKVLine("Group ID", groupID, tui.CDim, tui.Width))
		}
		fmt.Println(tui.RenderEmptyLine())

		if groupStatus == "NONE" || groupStatus == "" {
			fmt.Println(tui.RenderBoxDivider("OPSI", tui.Width))
			l1 := "  [ 1] Cek Kelayakan Membuat Grup"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Cek Kelayakan Membuat Grup"), len(l1)))

			l2 := "  [ 2] Buat Grup Family Hub Baru"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("2", tui.CSuc)+tui.BgSpaces(1)+tui.CWhi("Buat Grup Family Hub Baru"), len(l2)))

			lBack := "  [00] Kembali ke Menu Utama"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

			fmt.Println(tui.RenderEmptyLine())
			fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
			fmt.Println()

			choice := tui.TuiInput("Pilihan")
			if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
				return
			} else if choice == "1" {
				tui.RenderAlert("Memeriksa kelayakan...", "info")
				elRes, err := familyhub.GetFamilyHubEligibility(idToken)
				if err == nil && elRes != nil {
					code, _ := elRes["code"].(string)
					msg, _ := elRes["message"].(string)
					if code == "000" {
						tui.RenderAlert("Nomor Anda MEMENUHI SYARAT (ELIGIBLE)!", "success")
					} else {
						tui.RenderAlert(fmt.Sprintf("Tidak memenuhi syarat: %s", msg), "warning")
					}
				} else {
					tui.RenderAlert("Gagal memeriksa kelayakan.", "error")
				}
				tui.TuiPause()
			} else if choice == "2" {
				gName := tui.TuiInput("Nama Grup (contoh: Keluarga Bahagia)")
				pName := tui.TuiInput("Nama Pengelola (contoh: Ayah)")
				mNumRaw := tui.TuiInput("Nomor Anggota Pertama (08123456789)")
				mNum := config.NormalizeMSISDN(mNumRaw)
				mName := tui.TuiInput("Nama Anggota Pertama")

				if gName != "" && pName != "" && mNum != "" && mName != "" {
					members := []map[string]string{{"msisdn": mNum, "name": mName}}
					tui.RenderAlert("Membuat grup Family Hub...", "info")
					res, err := familyhub.CreateFamilyHubGroup(idToken, accessToken, gName, pName, members)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal membuat grup: %v", err), "error")
					} else {
						tui.RenderAlert("Grup Family Hub berhasil dibuat!", "success")
					}
					tui.TuiPause()
				}
			}
		} else {
			// ACTIVE or member of group
			memRes, _ := familyhub.GetFamilyHubMembers(idToken, groupID)
			var members []any
			if memRes != nil {
				if d, ok := memRes["data"].(map[string]any); ok {
					if m, ok := d["members"].([]any); ok {
						members = m
					}
				}
			}

			fmt.Println(tui.RenderBoxDivider(fmt.Sprintf("ANGGOTA GRUP (%d)", len(members)), tui.Width))
			for idx, m := range members {
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
			fmt.Println(tui.RenderBoxDivider("OPSI FAMILY HUB", tui.Width))

			l1 := "  [ 1] Undang Anggota Baru"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("1", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Undang Anggota Baru"), len(l1)))

			l2 := "  [ 2] Hapus Anggota dari Grup"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("2", tui.CDan)+tui.BgSpaces(1)+tui.CWhi("Hapus Anggota dari Grup"), len(l2)))

			l3 := "  [ 3] Ubah Nama Grup"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("3", tui.CPri)+tui.BgSpaces(1)+tui.CWhi("Ubah Nama Grup"), len(l3)))

			l4 := "  [ 4] Keluar dari Grup"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("4", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Keluar dari Grup"), len(l4)))

			l5 := "  [ 5] Bubarkan Grup"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("5", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Bubarkan Grup"), len(l5)))

			lBack := "  [00] Kembali ke Menu Utama"
			fmt.Println(tui.PLine(tui.BgSpaces(2)+tui.Badge("00", tui.CDan)+tui.BgSpaces(1)+tui.CDan("Kembali ke Menu Utama"), len(lBack)))

			fmt.Println(tui.RenderEmptyLine())
			fmt.Println(tui.RenderBoxBottom("Pilih opsi", tui.Width))
			fmt.Println()

			choice := tui.TuiInput("Pilihan")
			if choice == "00" || choice == "q" || choice == "exit" || choice == "" {
				return
			} else if choice == "1" {
				rawM := tui.TuiInput("Nomor anggota baru (contoh: 08123456789)")
				msisdn := config.NormalizeMSISDN(rawM)
				name := tui.TuiInput("Nama Anggota")
				if msisdn != "" && name != "" {
					members := []map[string]string{{"msisdn": msisdn, "name": name}}
					tui.RenderAlert("Mengirim undangan...", "info")
					res, err := familyhub.InviteFamilyHubMembers(idToken, accessToken, groupID, members)
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
				if err == nil && idx >= 1 && idx <= len(members) {
					target := members[idx-1].(map[string]any)
					mID, _ := target["member_id"].(string)
					if tui.TuiConfirm("Yakin ingin menghapus anggota ini?", false) {
						tui.RenderAlert("Menghapus anggota...", "info")
						res, err := familyhub.RemoveFamilyHubMember(idToken, accessToken, groupID, mID, len(members) == 1)
						if err != nil || res == nil {
							tui.RenderAlert(fmt.Sprintf("Gagal menghapus: %v", err), "error")
						} else {
							tui.RenderAlert("Anggota berhasil dihapus!", "success")
						}
						tui.TuiPause()
					}
				}
			} else if choice == "3" {
				newName := tui.TuiInput("Nama baru untuk grup")
				if newName != "" {
					tui.RenderAlert("Memperbarui nama grup...", "info")
					res, err := familyhub.ChangeFamilyHubGroupName(idToken, accessToken, groupID, newName)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal mengubah nama grup: %v", err), "error")
					} else {
						tui.RenderAlert("Nama grup berhasil diubah!", "success")
					}
					tui.TuiPause()
				}
			} else if choice == "4" {
				if tui.TuiConfirm("Yakin ingin keluar dari grup Family Hub ini?", false) {
					tui.RenderAlert("Memproses keluar grup...", "info")
					res, err := familyhub.LeaveFamilyHubGroup(idToken, accessToken, groupID, activeUser.SubscriberID)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal keluar grup: %v", err), "error")
					} else {
						tui.RenderAlert("Berhasil keluar dari grup!", "success")
					}
					tui.TuiPause()
				}
			} else if choice == "5" {
				if tui.TuiConfirm("PERINGATAN: Yakin ingin membubarkan grup Family Hub ini?", false) {
					tui.RenderAlert("Membubarkan grup...", "info")
					res, err := familyhub.DisbandFamilyHubGroup(idToken, accessToken, groupID)
					if err != nil || res == nil {
						tui.RenderAlert(fmt.Sprintf("Gagal membubarkan grup: %v", err), "error")
					} else {
						tui.RenderAlert("Grup Family Hub berhasil dibubarkan!", "success")
					}
					tui.TuiPause()
				}
			}
		}
	}
}
