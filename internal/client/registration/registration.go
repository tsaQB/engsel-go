package registration

import (
	"github.com/tsaQB/engsel-go/internal/client"
)

func ValidatePUK(msisdn, puk string) (map[string]any, error) {
	path := "api/v8/infos/validate-puk"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"puk":           puk,
		"is_enc":        false,
		"msisdn":        msisdn,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, rawPayload, "", "POST")
}

func Dukcapil(msisdn, kk, nik string) (map[string]any, error) {
	path := "api/v8/auth/regist/dukcapil"
	rawPayload := map[string]any{
		"msisdn": msisdn,
		"kk":     kk,
		"nik":    nik,
		"lang":   "en",
	}
	return client.SendAPIRequest(path, rawPayload, "", "POST")
}
