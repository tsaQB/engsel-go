package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
)

func EnterSentryMode() {
	activeUser := AuthInstance.GetActiveUser()
	if activeUser == nil {
		fmt.Println("Sesi login tidak aktif. Silakan login terlebih dahulu.")
		return
	}

	_ = os.MkdirAll("sentry", 0755)
	logFileName := filepath.Join("sentry", fmt.Sprintf("sentry_log_%s.jsonl", time.Now().Format("20060102_150405")))

	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Gagal membuka file log sentry: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Println("Memasuki Sentry Mode. Tekan ENTER untuk keluar.")
	fmt.Printf("Log disimpan di: %s\n", logFileName)

	var stopped int32

	go func() {
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
		atomic.StoreInt32(&stopped, 1)
	}()

	idToken, _ := activeUser.Tokens["id_token"].(string)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if atomic.LoadInt32(&stopped) == 1 {
			fmt.Println("Keluar dari Sentry Mode...")
			break
		}

		balance, err := client.GetBalance(idToken)
		entry := map[string]any{
			"timestamp": time.Now().Format(time.RFC3339),
			"balance":   balance,
			"error":     "",
		}
		if err != nil {
			entry["error"] = err.Error()
		}

		entryBytes, _ := json.Marshal(entry)
		_, _ = file.Write(append(entryBytes, '\n'))
		fmt.Printf("[%s] Sentry ping: logged status.\n", time.Now().Format("15:04:05"))

		select {
		case <-ticker.C:
		}
	}
}
