package responses

type MitraResponse struct {
	ID             int     `json:"id"`
	FirebaseToken  string  `json:"firebase_token"`
	IsAutoBid      string  `json:"is_auto_bid"`
	AccountBalance float64 `json:"account_balance"`
	TotalHutang    float64 `json:"total_hutang"`
	Distance       float64 `json:"distance"`
	MoneyLeft      float64 `json:"money_left"`
	CountOrder     int     `json:"count_order"`
	CountOrderRepeat int   `json:"count_order_repeat"`
}

type GetNearestMitraProductionResponse struct {
	IsAvailableNextTime bool            `json:"is_available_next_time"`
	PayloadMitra        []MitraResponse `json:"payload_mitra"`
	InitRange           int             `json:"init_range"`
}
