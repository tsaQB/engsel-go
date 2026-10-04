package circle

import (
	"github.com/tsaQB/engsel-go/internal/client"
	"github.com/tsaQB/engsel-go/internal/crypto"
)

func GetGroupData(tokens map[string]any) (map[string]any, error) {
	path := "family-hub/api/v8/groups/status"
	payload := map[string]any{
		"is_enterprise": false,
		"lang":          "en",
	}
	idToken, _ := tokens["id_token"].(string)
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetGroupMembers(tokens map[string]any, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/members/info"
	payload := map[string]any{
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "en",
	}
	idToken, _ := tokens["id_token"].(string)
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func ValidateCircleMember(tokens map[string]any, msisdn string) (map[string]any, error) {
	path := "family-hub/api/v8/members/validate"
	encMsisdn := crypto.EncryptCircleMSISDN(msisdn)
	payload := map[string]any{
		"msisdn":        encMsisdn,
		"is_enterprise": false,
		"lang":          "en",
	}
	idToken, _ := tokens["id_token"].(string)
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func InviteCircleMember(tokens map[string]any, msisdn, name, groupID, memberIDParent string) (map[string]any, error) {
	path := "family-hub/api/v8/members/invite"
	encMsisdn := crypto.EncryptCircleMSISDN(msisdn)
	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)

	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"is_enterprise": false,
		"members": []any{
			map[string]any{
				"msisdn": encMsisdn,
				"name":   name,
			},
		},
		"lang":             "en",
		"member_id_parent": memberIDParent,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func RemoveCircleMember(tokens map[string]any, memberID, groupID, memberIDParent string, isLastMember bool) (map[string]any, error) {
	path := "family-hub/api/v8/members/remove"
	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)

	payload := map[string]any{
		"access_token":     accessToken,
		"group_id":         groupID,
		"is_enterprise":    false,
		"is_last_member":   isLastMember,
		"lang":             "en",
		"member_id":        memberID,
		"member_id_parent": memberIDParent,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func AcceptCircleInvitation(tokens map[string]any, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/members/accept"
	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)

	payload := map[string]any{
		"access_token":  accessToken,
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func CreateCircle(tokens map[string]any, parentName, groupName, memberMSISDN, memberName string) (map[string]any, error) {
	path := "family-hub/api/v8/groups/create"
	encMsisdn := crypto.EncryptCircleMSISDN(memberMSISDN)
	accessToken, _ := tokens["access_token"].(string)
	idToken, _ := tokens["id_token"].(string)

	payload := map[string]any{
		"access_token":  accessToken,
		"group_name":    groupName,
		"is_enterprise": false,
		"lang":          "en",
		"members": []any{
			map[string]any{
				"msisdn": encMsisdn,
				"name":   memberName,
			},
		},
		"parent_name": parentName,
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func SpendingTracker(tokens map[string]any, groupID string) (map[string]any, error) {
	path := "family-hub/api/v8/gamification/spending-tracker"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "en",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}

func GetBonusData(tokens map[string]any, parentSubsID, familyID string) (map[string]any, error) {
	path := "family-hub/api/v8/bonus/detail"
	idToken, _ := tokens["id_token"].(string)
	payload := map[string]any{
		"family_id":        familyID,
		"is_enterprise":    false,
		"lang":             "en",
		"parent_subs_id":   parentSubsID,
		"reward_type":      "GROUP_BONUS",
	}
	return client.SendAPIRequest(path, payload, idToken, "POST")
}
