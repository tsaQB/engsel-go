package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
)

type RefreshTokenEntry struct {
	Number           int64  `json:"number"`
	SubscriberID     string `json:"subscriber_id"`
	SubscriptionType string `json:"subscription_type"`
	RefreshToken     string `json:"refresh_token"`
	AxFP             string `json:"ax_fp,omitempty"`
}

type ActiveUser struct {
	Number           int64          `json:"number"`
	SubscriberID     string         `json:"subscriber_id"`
	SubscriptionType string         `json:"subscription_type"`
	Tokens           map[string]any `json:"tokens"`
}

type AuthService struct {
	mu             sync.Mutex
	RefreshTokens  []RefreshTokenEntry
	ActiveUser     *ActiveUser
	LastRefreshSec int64
}

var AuthInstance = &AuthService{}

func (a *AuthService) Init() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.LoadTokensLocked()
	a.LoadActiveNumberLocked()
	a.LastRefreshSec = time.Now().Unix()
}

func (a *AuthService) LoadTokensLocked() {
	if _, err := os.Stat("refresh-tokens.json"); os.IsNotExist(err) {
		_ = os.WriteFile("refresh-tokens.json", []byte("[]"), 0644)
		a.RefreshTokens = []RefreshTokenEntry{}
		return
	}

	data, err := os.ReadFile("refresh-tokens.json")
	if err != nil {
		return
	}

	var rts []RefreshTokenEntry
	if err := json.Unmarshal(data, &rts); err == nil {
		a.RefreshTokens = rts
	}
}

func (a *AuthService) WriteTokensToFileLocked() {
	data, err := json.MarshalIndent(a.RefreshTokens, "", "    ")
	if err == nil {
		_ = os.WriteFile("refresh-tokens.json", data, 0644)
	}
}

func (a *AuthService) WriteActiveNumberLocked() {
	if a.ActiveUser != nil {
		_ = os.WriteFile("active.number", []byte(fmt.Sprintf("%d", a.ActiveUser.Number)), 0644)
	} else {
		_ = os.Remove("active.number")
	}
}

func (a *AuthService) LoadActiveNumberLocked() {
	if data, err := os.ReadFile("active.number"); err == nil {
		str := strings.TrimSpace(string(data))
		if num, err := strconv.ParseInt(str, 10, 64); err == nil {
			a.SetActiveUserLocked(num)
		}
	}
}

func (a *AuthService) SetActiveUserLocked(number int64) bool {
	var rtEntry *RefreshTokenEntry
	for i := range a.RefreshTokens {
		if a.RefreshTokens[i].Number == number {
			rtEntry = &a.RefreshTokens[i]
			break
		}
	}

	if rtEntry == nil {
		return false
	}

	preferredFP := rtEntry.AxFP
	tokens, err := client.GetNewToken(rtEntry.RefreshToken, rtEntry.SubscriberID, preferredFP)
	if err != nil || tokens == nil {
		return false
	}

	idToken, _ := tokens["id_token"].(string)
	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)

	profileData, err := client.GetProfile(accessToken, idToken)
	if err != nil || profileData == nil {
		return false
	}

	prof, _ := profileData["profile"].(map[string]any)
	if prof == nil {
		prof = profileData
	}

	subID, _ := prof["subscriber_id"].(string)
	subType, _ := prof["subscription_type"].(string)
	if subType == "" {
		subType = "PREPAID"
	}

	a.ActiveUser = &ActiveUser{
		Number:           number,
		SubscriberID:     subID,
		SubscriptionType: subType,
		Tokens:           tokens,
	}

	rtEntry.SubscriberID = subID
	rtEntry.SubscriptionType = subType
	rtEntry.RefreshToken = refreshToken
	if usedFP, ok := tokens["_used_fp"].(string); ok && usedFP != "" {
		rtEntry.AxFP = usedFP
	}

	a.WriteTokensToFileLocked()
	a.WriteActiveNumberLocked()
	a.LastRefreshSec = time.Now().Unix()
	return true
}

func (a *AuthService) SetActiveUser(number int64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.SetActiveUserLocked(number)
}

func (a *AuthService) AddRefreshToken(number int64, refreshToken string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	tokens, err := client.GetNewToken(refreshToken, "", "")
	if err != nil || tokens == nil {
		return false
	}

	idToken, _ := tokens["id_token"].(string)
	accessToken, _ := tokens["access_token"].(string)
	newRefreshToken, _ := tokens["refresh_token"].(string)
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	profileData, err := client.GetProfile(accessToken, idToken)
	subID := ""
	subType := "PREPAID"
	if err == nil && profileData != nil {
		if prof, ok := profileData["profile"].(map[string]any); ok {
			subID, _ = prof["subscriber_id"].(string)
			subType, _ = prof["subscription_type"].(string)
		}
	}

	var existing *RefreshTokenEntry
	for i := range a.RefreshTokens {
		if a.RefreshTokens[i].Number == number {
			existing = &a.RefreshTokens[i]
			break
		}
	}

	usedFP, _ := tokens["_used_fp"].(string)
	if existing != nil {
		existing.RefreshToken = newRefreshToken
		existing.SubscriberID = subID
		existing.SubscriptionType = subType
		if usedFP != "" {
			existing.AxFP = usedFP
		}
	} else {
		a.RefreshTokens = append(a.RefreshTokens, RefreshTokenEntry{
			Number:           number,
			SubscriberID:     subID,
			SubscriptionType: subType,
			RefreshToken:     newRefreshToken,
			AxFP:             usedFP,
		})
	}

	a.WriteTokensToFileLocked()
	return a.SetActiveUserLocked(number)
}

func (a *AuthService) RemoveRefreshToken(number int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var updated []RefreshTokenEntry
	for _, rt := range a.RefreshTokens {
		if rt.Number != number {
			updated = append(updated, rt)
		}
	}
	a.RefreshTokens = updated
	a.WriteTokensToFileLocked()

	if a.ActiveUser != nil && a.ActiveUser.Number == number {
		a.ActiveUser = nil
		_ = os.Remove("active.number")
		if len(a.RefreshTokens) > 0 {
			a.SetActiveUserLocked(a.RefreshTokens[0].Number)
		}
	}
}

func (a *AuthService) RenewActiveUserTokenLocked() bool {
	if a.ActiveUser == nil {
		return false
	}

	var rtEntry *RefreshTokenEntry
	for i := range a.RefreshTokens {
		if a.RefreshTokens[i].Number == a.ActiveUser.Number {
			rtEntry = &a.RefreshTokens[i]
			break
		}
	}

	preferredFP := ""
	if rtEntry != nil {
		preferredFP = rtEntry.AxFP
	}

	currRefreshToken, _ := a.ActiveUser.Tokens["refresh_token"].(string)
	tokens, err := client.GetNewToken(currRefreshToken, a.ActiveUser.SubscriberID, preferredFP)
	if err != nil || tokens == nil {
		return false
	}

	a.ActiveUser.Tokens = tokens
	a.LastRefreshSec = time.Now().Unix()

	if rtEntry != nil {
		if rt, ok := tokens["refresh_token"].(string); ok && rt != "" {
			rtEntry.RefreshToken = rt
		}
		if usedFP, ok := tokens["_used_fp"].(string); ok && usedFP != "" {
			rtEntry.AxFP = usedFP
		}
		a.WriteTokensToFileLocked()
	}
	return true
}

func (a *AuthService) GetActiveUser() *ActiveUser {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ActiveUser == nil {
		if len(a.RefreshTokens) > 0 {
			a.SetActiveUserLocked(a.RefreshTokens[0].Number)
		}
		return a.ActiveUser
	}

	if time.Now().Unix()-a.LastRefreshSec > 300 {
		_ = a.RenewActiveUserTokenLocked()
	}

	return a.ActiveUser
}

func (a *AuthService) GetActiveTokens() map[string]any {
	u := a.GetActiveUser()
	if u == nil {
		return nil
	}
	return u.Tokens
}
