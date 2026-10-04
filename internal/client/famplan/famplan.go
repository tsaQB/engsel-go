package famplan

import (
	"github.com/tsaQB/engsel-go/internal/client"
)

func GetFamilyData(tokens map[string]any) (map[string]any, error) {
	path := "sharings/api/v8/family-plan/member-info"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"group_id":      0,
		"is_enterprise": false,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ValidateMSISDN(tokens map[string]any, msisdn string) (map[string]any, error) {
	path := "api/v8/auth/check-dukcapil"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"with_bizon":          true,
		"with_family_plan":    true,
		"is_enterprise":       false,
		"with_optimus":        true,
		"lang":                "en",
		"msisdn":              msisdn,
		"with_regist_status":  true,
		"with_enterprise":     true,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ChangeMember(tokens map[string]any, parentAlias, alias string, slotID int, familyMemberID, newMsisdn string) (map[string]any, error) {
	path := "sharings/api/v8/family-plan/change-member"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"parent_alias":     parentAlias,
		"is_enterprise":    false,
		"slot_id":          slotID,
		"alias":            alias,
		"family_member_id": familyMemberID,
		"lang":             "en",
		"msisdn":           newMsisdn,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func RemoveMember(tokens map[string]any, parentAlias string, slotID int, familyMemberID string) (map[string]any, error) {
	path := "sharings/api/v8/family-plan/delete-member"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"family_member_id": familyMemberID,
		"is_enterprise":    false,
		"lang":             "en",
		"parent_alias":     parentAlias,
		"slot_id":          slotID,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func SetQuotaLimit(tokens map[string]any, parentAlias, alias string, slotID int, familyMemberID string, allocatedQuota int64) (map[string]any, error) {
	path := "sharings/api/v8/family-plan/set-quota-limit"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"parent_alias":     parentAlias,
		"is_enterprise":    false,
		"slot_id":          slotID,
		"alias":            alias,
		"allocated_quota":  allocatedQuota,
		"family_member_id": familyMemberID,
		"lang":             "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}
