package service

import (
	"encoding/json"
	"os"
	"sync"
)

type BookmarkItem struct {
	FamilyCode   string `json:"family_code"`
	FamilyName   string `json:"family_name"`
	IsEnterprise bool   `json:"is_enterprise"`
	VariantName  string `json:"variant_name"`
	OptionName   string `json:"option_name"`
	Order        int    `json:"order"`
}

type BookmarkService struct {
	mu       sync.Mutex
	FilePath string
	Packages []BookmarkItem
}

var BookmarkInstance = &BookmarkService{
	FilePath: "bookmark.json",
}

func (b *BookmarkService) Init() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Load()
}

func (b *BookmarkService) Load() {
	if _, err := os.Stat(b.FilePath); os.IsNotExist(err) {
		_ = os.WriteFile(b.FilePath, []byte("[]"), 0644)
		b.Packages = []BookmarkItem{}
		return
	}

	data, err := os.ReadFile(b.FilePath)
	if err != nil {
		return
	}

	var items []BookmarkItem
	if err := json.Unmarshal(data, &items); err == nil {
		b.Packages = items
	}
}

func (b *BookmarkService) Save() {
	data, err := json.MarshalIndent(b.Packages, "", "    ")
	if err == nil {
		_ = os.WriteFile(b.FilePath, data, 0644)
	}
}

func (b *BookmarkService) GetBookmarks() []BookmarkItem {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Load()
	return b.Packages
}

func (b *BookmarkService) AddBookmark(familyCode, familyName string, isEnterprise bool, variantName, optionName string, order int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, p := range b.Packages {
		if p.FamilyCode == familyCode && p.VariantName == variantName && p.Order == order && p.IsEnterprise == isEnterprise {
			return false // Already exists
		}
	}

	b.Packages = append(b.Packages, BookmarkItem{
		FamilyCode:   familyCode,
		FamilyName:   familyName,
		IsEnterprise: isEnterprise,
		VariantName:  variantName,
		OptionName:   optionName,
		Order:        order,
	})
	b.Save()
	return true
}

func (b *BookmarkService) RemoveBookmark(familyCode string, isEnterprise bool, variantName string, order int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	var updated []BookmarkItem
	removed := false
	for _, p := range b.Packages {
		if p.FamilyCode == familyCode && p.VariantName == variantName && p.Order == order && p.IsEnterprise == isEnterprise {
			removed = true
			continue
		}
		updated = append(updated, p)
	}

	if removed {
		b.Packages = updated
		b.Save()
	}
	return removed
}
