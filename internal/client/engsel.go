package client

import (
	"fmt"

	"github.com/tsaQB/engsel-go/internal/config"
)

func GetProfile(accessToken, idToken string) (map[string]any, error) {
	path := "api/v8/profile"
	rawPayload := map[string]any{
		"access_token":  accessToken,
		"app_version":   config.Cfg.AppVersion,
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetBalance(idToken string) (map[string]any, error) {
	path := "api/v8/packages/balance-and-credit"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	if data, ok := res["data"].(map[string]any); ok {
		if bal, ok := data["balance"].(map[string]any); ok {
			return bal, nil
		}
	}
	return nil, fmt.Errorf("balance data not found in response")
}

func GetFamily(idToken, familyCode string, isEnterprise bool) (map[string]any, error) {
	path := "api/v8/packages/family"
	rawPayload := map[string]any{
		"family_code":   familyCode,
		"is_enterprise": isEnterprise,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetFamilies(idToken, packageCategoryCode string) (map[string]any, error) {
	path := "api/v8/packages/families"
	rawPayload := map[string]any{
		"is_enterprise":         false,
		"lang":                  "en",
		"package_category_code": packageCategoryCode,
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetPackage(idToken, packageOptionCode string) (map[string]any, error) {
	path := "api/v8/packages/option"
	rawPayload := map[string]any{
		"is_enterprise":       false,
		"lang":                "en",
		"package_option_code": packageOptionCode,
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetAddons(idToken, packageOptionCode string) (map[string]any, error) {
	path := "api/v8/packages/addons"
	rawPayload := map[string]any{
		"is_enterprise":       false,
		"lang":                "en",
		"package_option_code": packageOptionCode,
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func InterceptPage(idToken, packageCode string, isEnterprise bool) (map[string]any, error) {
	path := "api/v8/xl-stores/intercept-page"
	rawPayload := map[string]any{
		"package_code":  packageCode,
		"is_enterprise": isEnterprise,
		"lang":          "en",
	}
	return SendAPIRequest(path, rawPayload, idToken, "POST")
}

func GetPackageDetails(idToken, packageOptionCode string) (map[string]any, error) {
	path := "api/v8/packages/details"
	rawPayload := map[string]any{
		"is_enterprise":       false,
		"lang":                "en",
		"package_option_code": packageOptionCode,
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetNotifications(idToken string) (map[string]any, error) {
	path := "api/v8/notifications"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}
	return SendAPIRequest(path, rawPayload, idToken, "POST")
}

func GetNotificationDetail(idToken, notificationID string) (map[string]any, error) {
	path := "api/v8/notifications/detail"
	rawPayload := map[string]any{
		"is_enterprise":   false,
		"lang":            "en",
		"notification_id": notificationID,
	}
	return SendAPIRequest(path, rawPayload, idToken, "POST")
}

func GetPendingTransaction(idToken string) (map[string]any, error) {
	path := "payments/api/v8/pending-transaction"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetTransactionHistory(idToken string) (map[string]any, error) {
	path := "payments/api/v8/transaction-history"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func GetTieringInfo(idToken string) (map[string]any, error) {
	path := "gamification/api/v8/loyalties/tiering/info"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := SendAPIRequest(path, rawPayload, idToken, "POST")
	if err != nil {
		return nil, err
	}
	if data, ok := res["data"].(map[string]any); ok {
		return data, nil
	}
	return res, nil
}

func Unsubscribe(idToken, subscriberID, optionCode string) (map[string]any, error) {
	path := "api/v8/packages/unsubscribe"
	rawPayload := map[string]any{
		"is_enterprise":       false,
		"lang":                "en",
		"package_option_code": optionCode,
		"subscriber_id":       subscriberID,
	}
	return SendAPIRequest(path, rawPayload, idToken, "POST")
}

func DashboardSegments(idToken string) (map[string]any, error) {
	path := "api/v8/dashboard/segments"
	rawPayload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}
	return SendAPIRequest(path, rawPayload, idToken, "POST")
}
