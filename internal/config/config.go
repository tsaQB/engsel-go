package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	BaseAPIURL        string
	BaseCIAMURL       string
	BasicAuth         string
	AxFpKey           string
	UA                string
	APIKey            string
	EncryptedFieldKey string
	XDataKey          string
	AxAPISigKey       string
	XAPIBaseSecret    string
	CircleMsisdnKey   string
	AppVersion        string
}

// Global active config
var Cfg Config

// LoadEnv loads environment variables from .env file if it exists.
func LoadEnv(filenames ...string) error {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Strip quotes if present
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}

			// Don't overwrite existing environment variables
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
	return nil
}

// InitConfig loads .env and populates the global Cfg struct.
func InitConfig() {
	_ = LoadEnv(".env")

	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "9.5.0"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = EnsureAPIKey()
	}

	Cfg = Config{
		BaseAPIURL:        strings.TrimRight(os.Getenv("BASE_API_URL"), "/"),
		BaseCIAMURL:       strings.TrimRight(os.Getenv("BASE_CIAM_URL"), "/"),
		BasicAuth:         os.Getenv("BASIC_AUTH"),
		AxFpKey:           os.Getenv("AX_FP_KEY"),
		UA:                os.Getenv("UA"),
		APIKey:            apiKey,
		EncryptedFieldKey: os.Getenv("ENCRYPTED_FIELD_KEY"),
		XDataKey:          os.Getenv("XDATA_KEY"),
		AxAPISigKey:       os.Getenv("AX_API_SIG_KEY"),
		XAPIBaseSecret:    os.Getenv("X_API_BASE_SECRET"),
		CircleMsisdnKey:   os.Getenv("CIRCLE_MSISDN_KEY"),
		AppVersion:        appVersion,
	}
}

func EnsureAPIKey() string {
	if _, err := os.Stat("api.key"); err == nil {
		data, err := os.ReadFile("api.key")
		if err == nil {
			key := strings.TrimSpace(string(data))
			if key != "" {
				return key
			}
		}
	}
	return "Noir1"
}

// NormalizeMSISDN standardizes Indonesian phone numbers to 628...
func NormalizeMSISDN(msisdn string) string {
	clean := strings.TrimSpace(msisdn)
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")

	if strings.HasPrefix(clean, "08") {
		return "628" + clean[2:]
	}
	if strings.HasPrefix(clean, "+628") {
		return "628" + clean[4:]
	}
	if strings.HasPrefix(clean, "+62") {
		return "62" + clean[3:]
	}
	return clean
}
