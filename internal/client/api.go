package client

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/crypto"
)

var HTTPClient = &http.Client{
	Timeout: 35 * time.Second,
}

// GenerateUUID generates a random UUID v4 string.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// SendAPIRequest sends an encrypted API request to BASE_API_URL and returns decrypted JSON as map or target type.
func SendAPIRequest(path string, payload any, idToken string, method ...string) (map[string]any, error) {
	reqMethod := "POST"
	if len(method) > 0 && method[0] != "" {
		reqMethod = method[0]
	}

	xSig, encryptedBody, err := crypto.EncryptSignXData(reqMethod, path, idToken, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt request: %w", err)
	}

	bodyBytes, err := json.Marshal(encryptedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encrypted body: %w", err)
	}

	xtime := encryptedBody["xtime"].(int64)
	sigTimeSec := xtime / 1000
	now := time.Now()

	baseURL := config.Cfg.BaseAPIURL
	fullURL := fmt.Sprintf("%s/%s", baseURL, strings.TrimPrefix(path, "/"))

	req, err := http.NewRequest(reqMethod, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	parsedURL, _ := url.Parse(baseURL)
	host := parsedURL.Host
	if host == "" {
		host = strings.Replace(baseURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
	}

	req.Header.Set("Host", host)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", config.Cfg.UA)
	req.Header.Set("X-Api-Key", config.Cfg.APIKey)
	if idToken != "" {
		req.Header.Set("Authorization", "Bearer "+idToken)
	}
	req.Header.Set("X-Hv", "v3")
	req.Header.Set("X-Signature-Time", fmt.Sprintf("%d", sigTimeSec))
	req.Header.Set("X-Signature", xSig)
	req.Header.Set("X-Request-Id", GenerateUUID())
	req.Header.Set("X-Request-At", crypto.JavaLikeTimestamp(now))
	req.Header.Set("X-Version-App", config.Cfg.AppVersion)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Response is expected to be encrypted JSON: {"xdata": "...", "xtime": ...}
	var respEncrypted struct {
		XData string `json:"xdata"`
		XTime any    `json:"xtime"`
	}
	if err := json.Unmarshal(respBytes, &respEncrypted); err != nil {
		// If not encrypted, return raw body mapped as error/message
		var rawMap map[string]any
		if json.Unmarshal(respBytes, &rawMap) == nil {
			return rawMap, nil
		}
		return map[string]any{"raw": string(respBytes), "status_code": resp.StatusCode}, nil
	}

	var xtimeVal int64
	switch v := respEncrypted.XTime.(type) {
	case float64:
		xtimeVal = int64(v)
	case int64:
		xtimeVal = v
	case json.Number:
		xtimeVal, _ = v.Int64()
	case string:
		fmt.Sscanf(v, "%d", &xtimeVal)
	}

	decryptedPlain, err := crypto.DecryptXData(respEncrypted.XData, xtimeVal)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt response: %w (raw response: %s)", err, string(respBytes))
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(decryptedPlain), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted JSON: %w (plaintext: %s)", err, decryptedPlain)
	}

	return result, nil
}
