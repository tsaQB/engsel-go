package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tsaQB/engsel-go/internal/config"
)

type DeviceInfo struct {
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	Lang           string  `json:"lang"`
	Resolution     string  `json:"resolution"`
	TzShort        string  `json:"tz_short"`
	IP             string  `json:"ip"`
	FontScale      float64 `json:"font_scale"`
	AndroidRelease string  `json:"android_release"`
	MSISDN         string  `json:"msisdn"`
}

// DeriveIV derives a 16-byte IV from sha256 of the millisecond timestamp as string.
func DeriveIV(xtimeMs int64) []byte {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d", xtimeMs)))
	hexStr := hex.EncodeToString(h[:])
	return []byte(hexStr[:16])
}

// PKCS7Pad pads data to a multiple of blockSize using PKCS#7.
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7Unpad removes PKCS#7 padding.
func PKCS7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("invalid padding size")
	}
	padLen := int(data[length-1])
	if padLen == 0 || padLen > blockSize || padLen > length {
		return nil, errors.New("invalid padding value")
	}
	for i := length - padLen; i < length; i++ {
		if data[i] != byte(padLen) {
			return nil, errors.New("invalid padding content")
		}
	}
	return data[:length-padLen], nil
}

// EncryptAESCBC encrypts plaintext with AES-CBC and given key and iv.
func EncryptAESCBC(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := PKCS7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

// DecryptAESCBC decrypts ciphertext with AES-CBC and unpads it.
func DecryptAESCBC(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	return PKCS7Unpad(plaintext, aes.BlockSize)
}

// EncryptXData encrypts plaintext with XDATA_KEY and timestamp-derived IV, returning urlsafe base64.
func EncryptXData(plaintext string, xtimeMs int64) (string, error) {
	key := []byte(config.Cfg.XDataKey)
	iv := DeriveIV(xtimeMs)
	ct, err := EncryptAESCBC(key, iv, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(ct), nil
}

// DecryptXData decrypts urlsafe base64 xdata using XDATA_KEY and timestamp-derived IV.
func DecryptXData(xdata string, xtimeMs int64) (string, error) {
	key := []byte(config.Cfg.XDataKey)
	iv := DeriveIV(xtimeMs)

	// Add missing '=' padding if necessary
	pad := (4 - (len(xdata) % 4)) % 4
	xdataPadded := xdata + strings.Repeat("=", pad)

	ct, err := base64.URLEncoding.DecodeString(xdataPadded)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}

	pt, err := DecryptAESCBC(key, iv, ct)
	if err != nil {
		return "", fmt.Errorf("aes decrypt error: %w", err)
	}

	return string(pt), nil
}

// MakeXSignature generates HMAC-SHA512 signature for regular requests.
func MakeXSignature(idToken, method, path string, sigTimeSec int64) string {
	keyStr := fmt.Sprintf("%s;%s;%s;%s;%d", config.Cfg.XAPIBaseSecret, idToken, method, path, sigTimeSec)
	msgStr := fmt.Sprintf("%s;%d;", idToken, sigTimeSec)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeXSignaturePayment generates HMAC-SHA512 signature for settlement payments.
func MakeXSignaturePayment(accessToken string, sigTimeSec int64, packageCode, tokenPayment, paymentMethod, paymentFor, path string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", config.Cfg.XAPIBaseSecret, sigTimeSec, path, sigTimeSec)
	msgStr := fmt.Sprintf("%s;%s;%d;%s;%s;%s;", accessToken, tokenPayment, sigTimeSec, paymentFor, paymentMethod, packageCode)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeAxAPISignature generates HMAC-SHA256 signature for CIAM login requests.
func MakeAxAPISignature(tsForSign, contact, code, contactType string) string {
	keyBytes := []byte(config.Cfg.AxAPISigKey)
	preimage := fmt.Sprintf("%spassword%s%s%sopenid", tsForSign, contactType, contact, code)

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(preimage))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureBounty generates HMAC-SHA512 signature for bounty exchange.
func MakeXSignatureBounty(accessToken string, sigTimeSec int64, packageCode, tokenPayment string) string {
	path := "api/v8/personalization/bounties-exchange"
	keyStr := fmt.Sprintf("%s;%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", config.Cfg.XAPIBaseSecret, accessToken, sigTimeSec, path, sigTimeSec)
	msgStr := fmt.Sprintf("%s;%s;%d;%s;", accessToken, tokenPayment, sigTimeSec, packageCode)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureLoyalty generates HMAC-SHA512 signature for loyalty exchange.
func MakeXSignatureLoyalty(sigTimeSec int64, packageCode, tokenConfirmation, path string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;POST;%s;%d", config.Cfg.XAPIBaseSecret, sigTimeSec, path, sigTimeSec)
	msgStr := fmt.Sprintf("%s;%d;%s;", tokenConfirmation, sigTimeSec, packageCode)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureBountyAllotment generates HMAC-SHA512 signature for bounty allotment.
func MakeXSignatureBountyAllotment(sigTimeSec int64, packageCode, tokenConfirmation, path, destinationMsisdn string) string {
	keyStr := fmt.Sprintf("%s;%d#ae-hei_9Tee6he+Ik3Gais5=;%s;POST;%s;%d", config.Cfg.XAPIBaseSecret, sigTimeSec, destinationMsisdn, path, sigTimeSec)
	msgStr := fmt.Sprintf("%s;%d;%s;%s;", tokenConfirmation, sigTimeSec, destinationMsisdn, packageCode)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureBalanceAllotment generates HMAC-SHA512 signature for balance transfer.
func MakeXSignatureBalanceAllotment(accessToken, receiverMsisdn string, amount int, path string) string {
	if path == "" {
		path = "sharings/api/v8/balance/allotment"
	}
	keyStr := fmt.Sprintf("%s;%s;%s;POST;%s", config.Cfg.XAPIBaseSecret, accessToken, receiverMsisdn, path)
	msgStr := fmt.Sprintf("%s;%s;%d;", accessToken, receiverMsisdn, amount)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// MakeXSignatureBasic generates basic HMAC-SHA512 signature.
func MakeXSignatureBasic(method, path string, sigTimeSec int64) string {
	keyStr := fmt.Sprintf("%s;%s;%s;%d", config.Cfg.XAPIBaseSecret, method, path, sigTimeSec)
	msgStr := fmt.Sprintf("%d;en;", sigTimeSec)

	mac := hmac.New(sha512.New, []byte(keyStr))
	mac.Write([]byte(msgStr))
	return hex.EncodeToString(mac.Sum(nil))
}

// RandomHex returns n random bytes encoded as hex (2*n hex chars).
func RandomHex(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// EncryptCircleMSISDN encrypts MSISDN for circle operations.
func EncryptCircleMSISDN(msisdn string) string {
	key := []byte(config.Cfg.EncryptedFieldKey)
	ivHex := RandomHex(8) // 16 hex chars
	iv := []byte(ivHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}
	padded := PKCS7Pad([]byte(msisdn), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	ctB64 := base64.URLEncoding.EncodeToString(ciphertext)
	return ctB64 + ivHex
}

// DecryptCircleMSISDN decrypts encrypted circle MSISDN.
func DecryptCircleMSISDN(encryptedMsisdnB64 string) string {
	if len(encryptedMsisdnB64) <= 16 {
		return ""
	}
	ivAscii := encryptedMsisdnB64[len(encryptedMsisdnB64)-16:]
	b64Part := encryptedMsisdnB64[:len(encryptedMsisdnB64)-16]

	key := []byte(config.Cfg.EncryptedFieldKey)
	iv := []byte(ivAscii)

	pad := (4 - (len(b64Part) % 4)) % 4
	b64Part += strings.Repeat("=", pad)

	ct, err := base64.URLEncoding.DecodeString(b64Part)
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher(key)
	if err != nil || len(ct)%aes.BlockSize != 0 {
		return ""
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	pt := make([]byte, len(ct))
	mode.CryptBlocks(pt, ct)

	unpadded, err := PKCS7Unpad(pt, aes.BlockSize)
	if err != nil {
		return ""
	}
	return string(unpadded)
}

// BuildEncryptedField encrypts empty bytes with AES-CBC and appends 16-hex IV.
func BuildEncryptedField(ivHex16 string, urlsafeB64 bool) string {
	key := []byte(config.Cfg.EncryptedFieldKey)
	ivHex := ivHex16
	if ivHex == "" {
		ivHex = RandomHex(8)
	}
	iv := []byte(ivHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}
	pt := PKCS7Pad([]byte(""), aes.BlockSize)
	ct := make([]byte, len(pt))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ct, pt)

	if urlsafeB64 {
		return base64.URLEncoding.EncodeToString(ct) + ivHex
	}
	return base64.StdEncoding.EncodeToString(ct) + ivHex
}

// BuildFingerprintPlain formats device info into string.
func BuildFingerprintPlain(dev DeviceInfo) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%.1f|Android %s|%s",
		dev.Manufacturer, dev.Model, dev.Lang, dev.Resolution,
		dev.TzShort, dev.IP, dev.FontScale, dev.AndroidRelease, dev.MSISDN)
}

// AxFingerprint generates device fingerprint string.
func AxFingerprint(dev DeviceInfo, secretKey string) (string, error) {
	key := []byte(secretKey)
	iv := make([]byte, 16)
	pt := []byte(BuildFingerprintPlain(dev))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	padded := PKCS7Pad(pt, aes.BlockSize)
	ct := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ct, padded)

	return base64.StdEncoding.EncodeToString(ct), nil
}

// AxDeviceID derives device ID as MD5 hex string of fingerprint.
func AxDeviceID(fp string) string {
	h := md5.Sum([]byte(fp))
	return hex.EncodeToString(h[:])
}

// LoadAxFP loads ax.fp file or generates a default one.
func LoadAxFP() string {
	fpPath := "ax.fp"
	if data, err := os.ReadFile(fpPath); err == nil {
		content := strings.TrimSpace(string(data))
		if content != "" {
			return content
		}
	}

	// Generate random device
	r1, _ := rand.Int(rand.Reader, big.NewInt(9000))
	r2, _ := rand.Int(rand.Reader, big.NewInt(9000))
	dev := DeviceInfo{
		Manufacturer:   fmt.Sprintf("samsung%d", 1000+r1.Int64()),
		Model:          fmt.Sprintf("SM-N93%d", 1000+r2.Int64()),
		Lang:           "en",
		Resolution:     "720x1540",
		TzShort:        "GMT07:00",
		IP:             "192.169.69.69",
		FontScale:      1.0,
		AndroidRelease: "13",
		MSISDN:         "6281398370564",
	}

	newFP, err := AxFingerprint(dev, config.Cfg.AxFpKey)
	if err == nil && newFP != "" {
		_ = os.WriteFile(fpPath, []byte(newFP), 0644)
		return newFP
	}
	return ""
}

// LoadAllFPs loads all available fingerprints from known candidate locations.
func LoadAllFPs() []string {
	var fps []string
	primary := LoadAxFP()
	if primary != "" {
		fps = append(fps, primary)
	}

	candidates := []string{
		"ax.fp.bak",
		"ax.fp.me",
		filepath.Join(os.Getenv("HOME"), "me-cli/ax.fp"),
		filepath.Join(os.Getenv("HOME"), "me-cli/ax.fp.bak"),
	}

	for _, p := range candidates {
		if data, err := os.ReadFile(p); err == nil {
			content := strings.TrimSpace(string(data))
			if content != "" {
				found := false
				for _, f := range fps {
					if f == content {
						found = true
						break
					}
				}
				if !found {
					fps = append(fps, content)
				}
			}
		}
	}
	return fps
}

// JavaLikeTimestamp formats time matching Python: "2006-01-02T15:04:05.ms+07:00"
func JavaLikeTimestamp(t time.Time) string {
	ms := t.Nanosecond() / 10_000_000 // 2 digits: 00..99
	zone, _ := t.Zone()
	_, offset := t.Zone()
	offsetHours := offset / 3600
	offsetMins := (offset % 3600) / 60
	tzStr := fmt.Sprintf("%+03d:%02d", offsetHours, offsetMins)
	_ = zone
	return fmt.Sprintf("%s.%02d%s", t.Format("2006-01-02T15:04:05"), ms, tzStr)
}

// TsGmt7WithoutColon formats time as "2006-01-02T15:04:05.000+0700"
func TsGmt7WithoutColon(t time.Time) string {
	loc := time.FixedZone("WIB", 7*3600)
	t7 := t.In(loc)
	return t7.Format("2006-01-02T15:04:05.000-0700")
}

// EncryptSignXData encrypts payload and calculates x-signature.
func EncryptSignXData(method, path, idToken string, payload any) (xSig string, encryptedBody map[string]any, err error) {
	plainBody, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	xtime := time.Now().UnixMilli()
	xdata, err := EncryptXData(string(plainBody), xtime)
	if err != nil {
		return "", nil, err
	}

	sigTimeSec := xtime / 1000
	xSig = MakeXSignature(idToken, method, path, sigTimeSec)

	encryptedBody = map[string]any{
		"xdata": xdata,
		"xtime": xtime,
	}
	return xSig, encryptedBody, nil
}
