package dtos

type NewsRequest struct {
	CreatorID    string `json:"creator_id"`
	CreatorName  string `json:"creator_name"`
	NewsTitle    string `json:"news_title" validate:"required"`
	NewsBody     string `json:"news_body" validate:"required"`
	NewsType     string `json:"news_type" validate:"required,oneof=Suberes Update News"`
	Narasumber   string `json:"narasumber" validate:"required"`
	IsBroadcast  string `json:"is_broadcast" validate:"required,oneof=0 1"`
	TimezoneCode string `json:"timezone_code" validate:"required"`
}
