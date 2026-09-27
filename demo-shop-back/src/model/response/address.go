package response

type AddressResp struct {
	AddressId     int64  `json:"address_id"`
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	PostalCode    string `json:"postal_code"`
	IsDefault     bool   `json:"is_default"`
	AddressTag    string `json:"address_tag"`
	CreatedAt     string `json:"created_at"`
}
