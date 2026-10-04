package familyhub

import (
	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/crypto"
)

func EnsureEncryptedMSISDN(msisdn string) string {
	if msisdn == "" {
		return ""
	}
	if len(msisdn) > 20 {
		return msisdn
	}
	return crypto.EncryptCircleMSISDN(msisdn)
}

func GetFamilyHubEligibility(idToken string) (map[string]any, error) {
	path := "family-hub/api/v8/check-eligibility"
	payload := map[string]any{
		"type":          "CREATE_GROUP",
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetFamilyHubStatus(idToken string) (map[string]any, error) {
	path := "family-hub/api/v8/groups/status"
	payload := map[string]any{
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetFamilyHubMembers(idToken, groupID string) (map[string]any, error) {
	path := "sharings/api/v8/family-hub/members"
	payload := map[string]any{
		"family_id":          groupID,
		"is_get_from_cache":  false,
		"is_enterprise":      false,
		"lang":               "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func CreateFamilyHubGroup(idToken, accessToken, groupName, parentName string, members []map[string]string) (map[string]any, error) {
	path := "family-hub/api/v8/groups/create"
	var encMembers []any
	for _, m := range members {
		encMembers = append(encMembers, map[string]any{
			"msisdn": EnsureEncryptedMSISDN(m["msisdn"]),
			"name":   m["name"],
		})
	}

	payload := map[string]any{
		"access_token":  accessToken,
		"group_name":    groupName,
		"parent_name":   parentName,
		"members":       encMembers,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func InviteFamilyHubMembers(idToken, accessToken, groupID string, members []map[string]string) (map[string]any, error) {
	path := "family-hub/api/v8/members/invite"
	var encMembers []any
	for _, m := range members {
		encMembers = append(encMembers, map[string]any{
			"msisdn": EnsureEncryptedMSISDN(m["msisdn"]),
			"name":   m["name"],
		})
	}

	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"members":       encMembers,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func RemoveFamilyHubMember(idToken, accessToken, groupID, memberID string, isLastMember bool) (map[string]any, error) {
	path := "family-hub/api/v8/members/remove"
	payload := map[string]any{
		"access_token":   accessToken,
		"group_id":       groupID,
		"member_id":      memberID,
		"is_last_member": isLastMember,
		"is_enterprise":  false,
		"lang":           "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func LeaveFamilyHubGroup(idToken, accessToken, groupID, memberID string) (map[string]any, error) {
	path := "family-hub/api/v8/members/leave"
	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"member_id":      memberID,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func DisbandFamilyHubGroup(idToken, accessToken, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/groups/disband"
	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ChangeFamilyHubGroupName(idToken, accessToken, groupID, newName string) (map[string]any, error) {
	path := "family-hub/api/v8/groups/change-name"
	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"group_name":    newName,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ChangeFamilyHubMemberName(idToken, accessToken, groupID, memberID, newName string) (map[string]any, error) {
	path := "family-hub/api/v8/members/change-name"
	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"member_id":      memberID,
		"name":          newName,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func AcceptFamilyHubInvitation(idToken, accessToken, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/members/accept"
	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetFamilyHubActivities(idToken, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/activities"
	payload := map[string]any{
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetFamilyHubPackages(idToken, accessToken string) (map[string]any, error) {
	path := "family-hub/api/v8/packages"
	payload := map[string]any{
		"access_token":  accessToken,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ValidateFamilyHubMember(idToken, msisdn string) (map[string]any, error) {
	path := "family-hub/api/v8/members/validate"
	encMsisdn := EnsureEncryptedMSISDN(msisdn)
	payload := map[string]any{
		"msisdn":        encMsisdn,
		"is_enterprise": false,
		"lang":          "id",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}
