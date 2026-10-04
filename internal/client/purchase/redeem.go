package purchase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/crypto"
)

func SettlementBounty(
	tokens map[string]any,
	tokenConfirmation string,
	tsToSign int64,
	paymentTarget string,
	price int,
) (map[string]any, error) {
	idToken, _ := tokens["id_token"].(string)
	accessToken, _ := tokens["access_token"].(string)

	path := "api/v8/personalization/bounties-exchange"
	settlementPayload := map[string]any{
		"total_discount":         0,
		"is_enterprise":          false,
		"payment_token":          "",
		"token_payment":          "",
		"activated_autobuy_code": "",
		"cc_payment_type":        "",
		"is_myxl_wallet":         false,
		"pin":                    "",
		"ewallet_promo_id":       "",
		"members":                []any{},
		"total_fee":              0,
		"fingerprint":            "",
		"autobuy_threshold_setting": map[string]any{
			"label": "",
			"type":  "",
			"value": 0,
		},
		"is_use_point":        false,
		"lang":                "en",
		"payment_method":      "BALANCE",
		"timestamp":           tsToSign,
		"points_gained":       0,
		"can_trigger_rating":  false,
		"akrab_members":       []any{},
		"akrab_parent_alias":  "",
		"referral_unique_code": "",
		"coupon":              "",
		"payment_for":         "REDEEM_VOUCHER",
		"with_upsell":         false,
		"topup_number":        "",
		"stage_token":         "",
		"authentication_id":   "",
		"encrypted_payment_token": crypto.BuildEncryptedField("", true),
		"token":               "",
		"token_confirmation":  tokenConfirmation,
		"access_token":        accessToken,
		"wallet_number":       "",
		"encrypted_authentication_id": crypto.BuildEncryptedField("", true),
		"additional_data": map[string]any{
			"original_price":           0,
			"is_spend_limit_temporary": false,
			"migration_type":           "",
			"akrab_m2m_group_id":       "",
			"spend_limit_amount":       0,
			"is_spend_limit":           false,
			"mission_id":               "",
			"tax":                      0,
			"benefit_type":             "",
			"quota_bonus":              0,
			"cashtag":                  "",
			"is_family_plan":           false,
			"combo_details":            []any{},
			"is_switch_plan":           false,
			"discount_recurring":       0,
			"is_akrab_m2m":             false,
			"balance_type":             "",
			"has_bonus":                false,
			"discount_promo":           0,
		},
		"total_amount":     0,
		"is_using_autobuy": false,
		"items": []any{
			map[string]any{
				"item_code":     paymentTarget,
				"product_type":  "",
				"item_price":    price,
				"item_name":     "",
				"tax":           0,
				"payment_target": paymentTarget,
			},
		},
	}

	bodyBytes, err := json.Marshal(settlementPayload)
	if err != nil {
		return nil, err
	}

	xtime := time.Now().UnixMilli()
	xdata, err := crypto.EncryptXData(string(bodyBytes), xtime)
	if err != nil {
		return nil, err
	}

	sigTimeSec := xtime / 1000
	xRequestedAt := time.Unix(sigTimeSec, 0)

	xSig := crypto.MakeXSignatureBounty(
		accessToken,
		sigTimeSec,
		paymentTarget,
		"",
	)

	encryptedBody := map[string]any{
		"xdata": xdata,
		"xtime": xtime,
	}
	encBodyBytes, _ := json.Marshal(encryptedBody)

	baseURL := config.Cfg.BaseAPIURL
	fullURL := fmt.Sprintf("%s/%s", baseURL, path)

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(encBodyBytes))
	if err != nil {
		return nil, err
	}

	parsedURL, _ := url.Parse(baseURL)
	host := parsedURL.Host
	if host == "" {
		host = strings.Replace(baseURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
	}

	req.Header.Set("host", host)
	req.Header.Set("content-type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", config.Cfg.UA)
	req.Header.Set("x-api-key", config.Cfg.APIKey)
	req.Header.Set("authorization", "Bearer "+idToken)
	req.Header.Set("x-hv", "v3")
	req.Header.Set("x-signature-time", fmt.Sprintf("%d", sigTimeSec))
	req.Header.Set("x-signature", xSig)
	req.Header.Set("x-request-id", client.GenerateUUID())
	req.Header.Set("x-request-at", crypto.JavaLikeTimestamp(xRequestedAt))
	req.Header.Set("x-version-app", config.Cfg.AppVersion)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respEncrypted struct {
		XData string `json:"xdata"`
		XTime any    `json:"xtime"`
	}
	if err := json.Unmarshal(respBytes, &respEncrypted); err != nil {
		var rawMap map[string]any
		_ = json.Unmarshal(respBytes, &rawMap)
		return rawMap, nil
	}

	var respXtime int64
	switch v := respEncrypted.XTime.(type) {
	case float64:
		respXtime = int64(v)
	case int64:
		respXtime = v
	}

	decryptedPlain, err := crypto.DecryptXData(respEncrypted.XData, respXtime)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bounty response: %w", err)
	}

	var result map[string]any
	_ = json.Unmarshal([]byte(decryptedPlain), &result)
	return result, nil
}

func SettlementLoyalty(
	tokens map[string]any,
	tokenConfirmation string,
	tsToSign int64,
	paymentTarget string,
	price int,
) (map[string]any, error) {
	idToken, _ := tokens["id_token"].(string)

	path := "gamification/api/v8/loyalties/tiering/exchange"
	settlementPayload := map[string]any{
		"item_code":          paymentTarget,
		"amount":             0,
		"partner":            "",
		"is_enterprise":      false,
		"item_name":          "",
		"lang":               "en",
		"points":             price,
		"timestamp":          tsToSign,
		"token_confirmation": tokenConfirmation,
	}

	bodyBytes, err := json.Marshal(settlementPayload)
	if err != nil {
		return nil, err
	}

	xtime := time.Now().UnixMilli()
	xdata, err := crypto.EncryptXData(string(bodyBytes), xtime)
	if err != nil {
		return nil, err
	}

	sigTimeSec := xtime / 1000
	xRequestedAt := time.Unix(sigTimeSec, 0)

	xSig := crypto.MakeXSignatureLoyalty(
		tsToSign,
		paymentTarget,
		tokenConfirmation,
		path,
	)

	encryptedBody := map[string]any{
		"xdata": xdata,
		"xtime": xtime,
	}
	encBodyBytes, _ := json.Marshal(encryptedBody)

	baseURL := config.Cfg.BaseAPIURL
	fullURL := fmt.Sprintf("%s/%s", baseURL, path)

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(encBodyBytes))
	if err != nil {
		return nil, err
	}

	parsedURL, _ := url.Parse(baseURL)
	host := parsedURL.Host
	if host == "" {
		host = strings.Replace(baseURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
	}

	req.Header.Set("host", host)
	req.Header.Set("content-type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", config.Cfg.UA)
	req.Header.Set("x-api-key", config.Cfg.APIKey)
	req.Header.Set("authorization", "Bearer "+idToken)
	req.Header.Set("x-hv", "v3")
	req.Header.Set("x-signature-time", fmt.Sprintf("%d", sigTimeSec))
	req.Header.Set("x-signature", xSig)
	req.Header.Set("x-request-id", client.GenerateUUID())
	req.Header.Set("x-request-at", crypto.JavaLikeTimestamp(xRequestedAt))
	req.Header.Set("x-version-app", config.Cfg.AppVersion)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respEncrypted struct {
		XData string `json:"xdata"`
		XTime any    `json:"xtime"`
	}
	if err := json.Unmarshal(respBytes, &respEncrypted); err != nil {
		var rawMap map[string]any
		_ = json.Unmarshal(respBytes, &rawMap)
		return rawMap, nil
	}

	var respXtime int64
	switch v := respEncrypted.XTime.(type) {
	case float64:
		respXtime = int64(v)
	case int64:
		respXtime = v
	}

	decryptedPlain, err := crypto.DecryptXData(respEncrypted.XData, respXtime)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt loyalty response: %w", err)
	}

	var result map[string]any
	_ = json.Unmarshal([]byte(decryptedPlain), &result)
	return result, nil
}

func BountyAllotment(
	tokens map[string]any,
	tsToSign int64,
	destinationMsisdn string,
	itemName string,
	itemCode string,
	tokenConfirmation string,
) (map[string]any, error) {
	idToken, _ := tokens["id_token"].(string)

	path := "gamification/api/v8/loyalties/tiering/bounties-allotment"
	settlementPayload := map[string]any{
		"destination_msisdn": destinationMsisdn,
		"item_code":          itemCode,
		"is_enterprise":      false,
		"item_name":          itemName,
		"lang":               "en",
		"timestamp":          time.Now().Unix(),
		"token_confirmation": tokenConfirmation,
	}

	bodyBytes, err := json.Marshal(settlementPayload)
	if err != nil {
		return nil, err
	}

	xtime := time.Now().UnixMilli()
	xdata, err := crypto.EncryptXData(string(bodyBytes), xtime)
	if err != nil {
		return nil, err
	}

	sigTimeSec := xtime / 1000
	xRequestedAt := time.Unix(sigTimeSec, 0)

	xSig := crypto.MakeXSignatureBountyAllotment(
		tsToSign,
		itemCode,
		tokenConfirmation,
		path,
		destinationMsisdn,
	)

	encryptedBody := map[string]any{
		"xdata": xdata,
		"xtime": xtime,
	}
	encBodyBytes, _ := json.Marshal(encryptedBody)

	baseURL := config.Cfg.BaseAPIURL
	fullURL := fmt.Sprintf("%s/%s", baseURL, path)

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(encBodyBytes))
	if err != nil {
		return nil, err
	}

	parsedURL, _ := url.Parse(baseURL)
	host := parsedURL.Host
	if host == "" {
		host = strings.Replace(baseURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
	}

	req.Header.Set("host", host)
	req.Header.Set("content-type", "application/json; charset=utf-8")
	req.Header.Set("user-agent", config.Cfg.UA)
	req.Header.Set("x-api-key", config.Cfg.APIKey)
	req.Header.Set("authorization", "Bearer "+idToken)
	req.Header.Set("x-hv", "v3")
	req.Header.Set("x-signature-time", fmt.Sprintf("%d", sigTimeSec))
	req.Header.Set("x-signature", xSig)
	req.Header.Set("x-request-id", client.GenerateUUID())
	req.Header.Set("x-request-at", crypto.JavaLikeTimestamp(xRequestedAt))
	req.Header.Set("x-version-app", config.Cfg.AppVersion)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respEncrypted struct {
		XData string `json:"xdata"`
		XTime any    `json:"xtime"`
	}
	if err := json.Unmarshal(respBytes, &respEncrypted); err != nil {
		var rawMap map[string]any
		_ = json.Unmarshal(respBytes, &rawMap)
		return rawMap, nil
	}

	var respXtime int64
	switch v := respEncrypted.XTime.(type) {
	case float64:
		respXtime = int64(v)
	case int64:
		respXtime = v
	}

	decryptedPlain, err := crypto.DecryptXData(respEncrypted.XData, respXtime)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bounty allotment response: %w", err)
	}

	var result map[string]any
	_ = json.Unmarshal([]byte(decryptedPlain), &result)
	return result, nil
}
