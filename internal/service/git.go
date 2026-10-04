package service

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	RepoOwner  = "tsaQB"
	RepoName   = "engsel-go"
	RepoBranch = "main"
)

type AtomFeed struct {
	Entry struct {
		ID string `xml:"id"`
	} `xml:"entry"`
}

func GetLocalCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetLatestCommitAtom() (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/commits/%s.atom", RepoOwner, RepoName, RepoBranch)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var feed AtomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return "", err
	}

	id := feed.Entry.ID
	if id == "" {
		return "", fmt.Errorf("empty entry id")
	}

	parts := strings.Split(id, "/")
	return parts[len(parts)-1], nil
}

func CheckForUpdates() bool {
	local := GetLocalCommit()
	if local == "" {
		return false
	}

	remote, err := GetLatestCommitAtom()
	if err != nil || remote == "" {
		return false
	}

	if local != remote {
		lShort := local
		if len(lShort) > 7 {
			lShort = lShort[:7]
		}
		rShort := remote
		if len(rShort) > 7 {
			rShort = rShort[:7]
		}
		fmt.Printf("⚠️  Versi baru tersedia (remote %s vs lokal %s).\n", rShort, lShort)
		fmt.Println("   Jalankan: git pull --rebase untuk memperbarui.")
		return true
	}
	return false
}
