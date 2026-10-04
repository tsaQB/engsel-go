package client

import (
	"bytes"
	"encoding/base64"
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

var ActiveFP = crypto.LoadAxFP()

func SetActiveFP(fp string) {
	if fp != "" {
		ActiveFP = fp
	}
}

func ValidateContact(contact string) bool {
	contact = config.NormalizeMSISDN(contact)
	if !strings.HasPrefix(contact, "628") || len(contact) > 14 {
		return false
	}
	return true
}

func GetOTP(contact string) (string, error) {
	contact = config.NormalizeMSISDN(contact)
	if !ValidateContact(contact) {
		return "", fmt.Errorf("invalid phone number")
	}

	rawURL := fmt.Sprintf("%s/realms/xl-ciam/auth/otp", config.Cfg.BaseCIAMURL)
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := reqURL.Query()
	q.Set("contact", contact)
	q.Set("contactType", "SMS")
	q.Set("alternateContact", "false")
	reqURL.RawQuery = q.Encode()

	now := time.Now()
	loc := time.FixedZone("WIB", 7*3600)
	nowGmt7 := now.In(loc)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return "", err
	}

	devID := crypto.AxDeviceID(ActiveFP)
	host := strings.Replace(config.Cfg.BaseCIAMURL, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Authorization", "Basic "+config.Cfg.BasicAuth)
	req.Header.Set("Ax-Device-Id", devID)
	req.Header.Set("Ax-Fingerprint", ActiveFP)
	req.Header.Set("Ax-Request-At", crypto.JavaLikeTimestamp(nowGmt7))
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Request-Id", GenerateUUID())
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("User-Agent", config.Cfg.UA)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send OTP request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var jsonBody map[string]any
	if err := json.Unmarshal(respBytes, &jsonBody); err != nil {
		return "", fmt.Errorf("invalid response from server: %s", string(respBytes))
	}

	if subID, ok := jsonBody["subscriber_id"].(string); ok && subID != "" {
		return subID, nil
	}

	errMsg := "Subscriber ID not found in response"
	if errVal, ok := jsonBody["error"].(string); ok {
		errMsg = errVal
	}
	return "", fmt.Errorf("%s (body: %s)", errMsg, string(respBytes))
}

func ExtendSession(subscriberID string, fp string) (string, error) {
	useFP := fp
	if useFP == "" {
		useFP = ActiveFP
	}
	devID := crypto.AxDeviceID(useFP)

	b64SubscriberID := base64.StdEncoding.EncodeToString([]byte(subscriberID))
	rawURL := fmt.Sprintf("%s/realms/xl-ciam/auth/extend-session", config.Cfg.BaseCIAMURL)
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := reqURL.Query()
	q.Set("contact", b64SubscriberID)
	q.Set("contactType", "DEVICEID")
	reqURL.RawQuery = q.Encode()

	now := time.Now()
	loc := time.FixedZone("WIB", 7*3600)
	nowGmt7 := now.In(loc)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return "", err
	}

	host := strings.Replace(config.Cfg.BaseCIAMURL, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Authorization", "Basic "+config.Cfg.BasicAuth)
	req.Header.Set("Ax-Device-Id", devID)
	req.Header.Set("Ax-Fingerprint", useFP)
	req.Header.Set("Ax-Request-At", crypto.JavaLikeTimestamp(nowGmt7))
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Request-Id", GenerateUUID())
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("User-Agent", config.Cfg.UA)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to extend session (%d): %s", resp.StatusCode, string(respBytes))
	}

	var data map[string]any
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return "", err
	}

	if d, ok := data["data"].(map[string]any); ok {
		if exCode, ok := d["exchange_code"].(string); ok {
			return exCode, nil
		}
	}
	return "", fmt.Errorf("exchange_code not found in response: %s", string(respBytes))
}

func SubmitOTP(contactType, contact, code string, fp string) (map[string]any, error) {
	useFP := fp
	if useFP == "" {
		useFP = ActiveFP
	}
	devID := crypto.AxDeviceID(useFP)

	finalContact := ""
	finalCode := ""

	if contactType == "SMS" {
		contact = config.NormalizeMSISDN(contact)
		if !ValidateContact(contact) {
			return nil, fmt.Errorf("invalid phone number")
		}
		finalContact = contact
		if len(code) != 6 {
			return nil, fmt.Errorf("invalid OTP code format (must be 6 digits)")
		}
		finalCode = code
	} else if contactType == "DEVICEID" {
		finalContact = base64.StdEncoding.EncodeToString([]byte(contact))
		finalCode = code
	} else {
		return nil, fmt.Errorf("unsupported contact type: %s", contactType)
	}

	rawURL := fmt.Sprintf("%s/realms/xl-ciam/protocol/openid-connect/token", config.Cfg.BaseCIAMURL)

	now := time.Now()
	loc := time.FixedZone("WIB", 7*3600)
	nowGmt7 := now.In(loc)

	tsForSign := crypto.TsGmt7WithoutColon(nowGmt7)
	tsHeader := crypto.TsGmt7WithoutColon(nowGmt7.Add(-5 * time.Minute))
	signature := crypto.MakeAxAPISignature(tsForSign, finalContact, code, contactType)

	payload := fmt.Sprintf("contactType=%s&code=%s&grant_type=password&contact=%s&scope=openid",
		url.QueryEscape(contactType), url.QueryEscape(finalCode), url.QueryEscape(finalContact))

	req, err := http.NewRequest("POST", rawURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Authorization", "Basic "+config.Cfg.BasicAuth)
	req.Header.Set("Ax-Api-Signature", signature)
	req.Header.Set("Ax-Device-Id", devID)
	req.Header.Set("Ax-Fingerprint", useFP)
	req.Header.Set("Ax-Request-At", tsHeader)
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Request-Id", GenerateUUID())
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", config.Cfg.UA)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonBody map[string]any
	if err := json.Unmarshal(respBytes, &jsonBody); err != nil {
		return nil, fmt.Errorf("invalid json: %s", string(respBytes))
	}

	if _, hasErr := jsonBody["error"]; hasErr {
		return nil, fmt.Errorf("submit OTP error: %s", string(respBytes))
	}

	return jsonBody, nil
}

func GetNewToken(refreshToken, subscriberID, preferredFP string) (map[string]any, error) {
	rawURL := fmt.Sprintf("%s/realms/xl-ciam/protocol/openid-connect/token", config.Cfg.BaseCIAMURL)

	allFPs := crypto.LoadAllFPs()
	var fpsToTry []string
	if preferredFP != "" {
		fpsToTry = append(fpsToTry, preferredFP)
	}
	for _, f := range allFPs {
		exists := false
		for _, p := range fpsToTry {
			if p == f {
				exists = true
				break
			}
		}
		if !exists {
			fpsToTry = append(fpsToTry, f)
		}
	}

	var lastError string

	for _, fp := range fpsToTry {
		devID := crypto.AxDeviceID(fp)
		now := time.Now()
		loc := time.FixedZone("WIB", 7*3600)
		axRequestAt := now.In(loc).Format("2006-01-02T15:04:05.000-0700")

		data := url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", refreshToken)

		req, err := http.NewRequest("POST", rawURL, strings.NewReader(data.Encode()))
		if err != nil {
			lastError = err.Error()
			continue
		}

		host := strings.Replace(config.Cfg.BaseCIAMURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)

		req.Header.Set("Host", host)
		req.Header.Set("ax-request-at", axRequestAt)
		req.Header.Set("ax-device-id", devID)
		req.Header.Set("ax-request-id", GenerateUUID())
		req.Header.Set("ax-request-device", "samsung")
		req.Header.Set("ax-request-device-model", "SM-N935F")
		req.Header.Set("ax-fingerprint", fp)
		req.Header.Set("authorization", "Basic "+config.Cfg.BasicAuth)
		req.Header.Set("user-agent", config.Cfg.UA)
		req.Header.Set("ax-substype", "PREPAID")
		req.Header.Set("content-type", "application/x-www-form-urlencoded")

		resp, err := HTTPClient.Do(req)
		if err != nil {
			lastError = err.Error()
			continue
		}

		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var body map[string]any
			if json.Unmarshal(respBytes, &body) == nil {
				if _, ok := body["id_token"]; ok {
					body["_used_fp"] = fp
					SetActiveFP(fp)
					return body, nil
				}
			}
		}

		if resp.StatusCode == http.StatusBadRequest {
			var errJSON map[string]any
			_ = json.Unmarshal(respBytes, &errJSON)

			errDesc, _ := errJSON["error_description"].(string)
			errMsg, _ := errJSON["error"].(string)

			canExtend := errDesc == "Session not active" ||
				strings.Contains(errMsg, "User Agent mismatch") ||
				strings.Contains(string(respBytes), "User Agent mismatch")

			if canExtend && subscriberID != "" {
				exchangeCode, extErr := ExtendSession(subscriberID, fp)
				if extErr == nil && exchangeCode != "" {
					extendResult, subErr := SubmitOTP("DEVICEID", subscriberID, exchangeCode, fp)
					if subErr == nil && extendResult != nil {
						if _, ok := extendResult["refresh_token"]; ok {
							extendResult["_used_fp"] = fp
							SetActiveFP(fp)
							return extendResult, nil
						}
					}
				}
			}
		}
		lastError = fmt.Sprintf("%d - %s", resp.StatusCode, string(respBytes))
	}

	return nil, fmt.Errorf("failed to refresh token: %s", lastError)
}

func GetAuthCode(accessToken, pin, msisdn string) (string, error) {
	rawURL := fmt.Sprintf("%s/ciam/auth/authorization-token/generate", config.Cfg.BaseCIAMURL)

	parsedURL, _ := url.Parse(config.Cfg.BaseCIAMURL)
	host := parsedURL.Host
	if host == "" {
		host = strings.Replace(config.Cfg.BaseCIAMURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
	}

	now := time.Now()
	loc := time.FixedZone("WIB", 7*3600)
	axRequestAt := now.In(loc).Format("2006-01-02T15:04:05.000-0700")

	pinB64 := base64.StdEncoding.EncodeToString([]byte(pin))
	bodyPayload := map[string]any{
		"pin":              pinB64,
		"transaction_type": "SHARE_BALANCE",
		"receiver_msisdn":  msisdn,
	}
	bodyBytes, _ := json.Marshal(bodyPayload)

	req, err := http.NewRequest("POST", rawURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	devID := crypto.AxDeviceID(ActiveFP)

	req.Header.Set("Host", host)
	req.Header.Set("Ax-Request-At", axRequestAt)
	req.Header.Set("Ax-Device-Id", devID)
	req.Header.Set("Ax-Request-Id", GenerateUUID())
	req.Header.Set("Ax-Request-Device", "samsung")
	req.Header.Set("Ax-Request-Device-Model", "SM-N935F")
	req.Header.Set("Ax-Fingerprint", ActiveFP)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", config.Cfg.UA)
	req.Header.Set("Ax-Substype", "PREPAID")
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get auth code (%d): %s", resp.StatusCode, string(respBytes))
	}

	var data map[string]any
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return "", err
	}

	if d, ok := data["data"].(map[string]any); ok {
		if authCode, ok := d["authorization_code"].(string); ok && authCode != "" {
			return authCode, nil
		}
	}

	return "", fmt.Errorf("authorization_code not found: %s", string(respBytes))
}
