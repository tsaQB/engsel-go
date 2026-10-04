package sharing

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

func BalanceAllotment(
	tokens map[string]any,
	stageToken string,
	receiverMsisdn string,
	amount int,
) (map[string]any, error) {
	idToken, _ := tokens["id_token"].(string)
	accessToken, _ := tokens["access_token"].(string)

	path := "sharings/api/v8/balance/allotment"
	allotmentPayload := map[string]any{
		"access_token":  accessToken,
		"receiver":      receiverMsisdn,
		"amount":        amount,
		"stage_token":   stageToken,
		"lang":          "en",
		"is_enterprise": false,
	}

	bodyBytes, err := json.Marshal(allotmentPayload)
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

	xSig := crypto.MakeXSignatureBalanceAllotment(
		accessToken,
		receiverMsisdn,
		amount,
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
		return nil, fmt.Errorf("failed to decrypt response: %w", err)
	}

	var result map[string]any
	_ = json.Unmarshal([]byte(decryptedPlain), &result)
	return result, nil
}
