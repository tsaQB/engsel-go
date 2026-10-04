package purchase

type PaymentItem struct {
	ItemCode          string `json:"item_code"`
	ProductType       string `json:"product_type"`
	ItemPrice         int    `json:"item_price"`
	ItemName          string `json:"item_name"`
	Tax               int    `json:"tax"`
	TokenConfirmation string `json:"token_confirmation"`
}
