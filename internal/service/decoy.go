package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tsaQB/engsel-go/internal/client"
)

type DecoyInfo struct {
	OptionCode    string `json:"option_code"`
	Price         int    `json:"price"`
	LastFetchedAt int64  `json:"last_fetched_at"`
}

type DecoyConfigFile struct {
	FamilyName    string `json:"family_name"`
	FamilyCode    string `json:"family_code"`
	IsEnterprise  bool   `json:"is_enterprise"`
	MigrationType string `json:"migration_type"`
	VariantCode   string `json:"variant_code"`
	OptionName    string `json:"option_name"`
	Order         int    `json:"order"`
	Price         int    `json:"price"`
}

type DecoyService struct {
	mu               sync.Mutex
	DecoyBasePath    string
	SubscriberID     string
	SubscriptionType string
	Prefix           string
	Decoys           map[string]DecoyInfo
}

var DecoyInstance = &DecoyService{
	DecoyBasePath: "decoy_data/decoy-",
	Prefix:        "default-",
	Decoys:        make(map[string]DecoyInfo),
}

func (d *DecoyService) CheckSubscriberChange() {
	activeUser := AuthInstance.GetActiveUser()
	if activeUser == nil {
		return
	}

	if d.SubscriberID != activeUser.SubscriberID {
		d.SubscriberID = activeUser.SubscriberID
		d.SubscriptionType = activeUser.SubscriptionType
		d.Decoys = make(map[string]DecoyInfo)

		prioTypes := map[string]bool{
			"PRIORITAS":  true,
			"PRIOHYBRID": true,
			"GO":         true,
		}
		if prioTypes[d.SubscriptionType] {
			d.Prefix = "prio-"
		} else {
			d.Prefix = "default-"
		}
	}
}

func (d *DecoyService) FetchDecoyData(decoyName string) {
	activeUser := AuthInstance.GetActiveUser()
	if activeUser == nil {
		return
	}

	filePath := d.DecoyBasePath + decoyName + ".json"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Try root decoy_data path
		filePath = filepath.Join("decoy_data", fmt.Sprintf("decoy-%s.json", decoyName))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var cfg DecoyConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	idToken, _ := activeUser.Tokens["id_token"].(string)

	// Fetch family and find the option code
	familyData, err := client.GetFamily(idToken, cfg.FamilyCode, cfg.IsEnterprise)
	if err != nil || familyData == nil {
		return
	}

	variants, _ := familyData["package_variants"].([]any)
	for _, v := range variants {
		vMap, _ := v.(map[string]any)
		if vMap == nil {
			continue
		}
		vCode, _ := vMap["package_variant_code"].(string)
		if vCode == cfg.VariantCode || cfg.VariantCode == "" {
			options, _ := vMap["package_options"].([]any)
			for _, opt := range options {
				optMap, _ := opt.(map[string]any)
				if optMap == nil {
					continue
				}
				orderVal := 0
				switch ov := optMap["order"].(type) {
				case float64:
					orderVal = int(ov)
				case int:
					orderVal = ov
				}
				if orderVal == cfg.Order {
					optCode, _ := optMap["package_option_code"].(string)
					d.Decoys[decoyName] = DecoyInfo{
						OptionCode:    optCode,
						Price:         cfg.Price,
						LastFetchedAt: time.Now().Unix(),
					}
					return
				}
			}
		}
	}
}

func (d *DecoyService) GetDecoy(paymentType string) *DecoyInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.CheckSubscriberChange()

	decoyName := d.Prefix + paymentType
	decoy, ok := d.Decoys[decoyName]
	if !ok || time.Now().Unix()-decoy.LastFetchedAt > 300 {
		d.FetchDecoyData(decoyName)
		decoy, ok = d.Decoys[decoyName]
	}

	if !ok {
		return nil
	}
	return &decoy
}
