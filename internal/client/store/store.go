package store

import (
	"github.com/tsaQB/engsel-go/internal/client"
)

func GetSegments(idToken string, isEnterprise bool) (map[string]any, error) {
	path := "api/v8/configs/store/segments"
	payload := map[string]any{
		"is_enterprise": isEnterprise,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetFamilyList(idToken string, subsType string, isEnterprise bool) (map[string]any, error) {
	path := "api/v8/xl-stores/options/search/family-list"
	payload := map[string]any{
		"is_enterprise": isEnterprise,
		"subs_type":     subsType,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetStorePackages(idToken string, subsType string, isEnterprise bool) (map[string]any, error) {
	path := "api/v9/xl-stores/options/search"
	payload := map[string]any{
		"is_enterprise": isEnterprise,
		"filters": []any{
			map[string]any{
				"unit":  "THOUSAND",
				"id":    "FIL_SEL_P",
				"type":  "PRICE",
				"items": []any{},
			},
			map[string]any{
				"unit":  "GB",
				"id":    "FIL_SEL_MQ",
				"type":  "DATA_TYPE",
				"items": []any{},
			},
			map[string]any{
				"unit": "PACKAGE_NAME",
				"id":   "FIL_PKG_N",
				"type": "PACKAGE_NAME",
				"items": []any{
					map[string]any{
						"id":    "",
						"label": "",
					},
				},
			},
			map[string]any{
				"unit":  "DAY",
				"id":    "FIL_SEL_V",
				"type":  "VALIDITY",
				"items": []any{},
			},
		},
		"substype":    subsType,
		"text_search": "",
		"lang":        "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetRedeemables(idToken string, isEnterprise bool) (map[string]any, error) {
	path := "api/v8/personalization/redeemables"
	payload := map[string]any{
		"is_enterprise": isEnterprise,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}
