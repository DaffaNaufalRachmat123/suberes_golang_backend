package dtos

type DashboardPayload struct {
	TodayOrderData       TodayOrderData       `json:"today_order_data"`
	TodayMitraData       TodayMitraData       `json:"today_mitra_data"`
	TodayTransactionData TodayTransactionData `json:"today_transaction_data"`
	TotalOrdersByMonth   []TotalOrdersByMonth `json:"total_orders_by_month"`
	OverviewMonth        []OverviewMonthRevenue `json:"overview_month"`
	OverviewWeek         []OverviewWeekRevenue  `json:"overview_week"`
	FusService           []FrequentlyUsedService `json:"fus_service"`
	MitraOrderToday      []MitraOrderToday      `json:"mitra_order_today"`
}

type TodayOrderData struct {
	Percentage float64 `json:"percentage"`
	TodayCount int64   `json:"today_count"`
}

type TodayMitraData struct {
	Percentage float64 `json:"percentage"`
	TodayCount int64   `json:"today_count"`
}

type TodayTransactionData struct {
	Percentage float64 `json:"percentage"`
	TodayCount int64   `json:"today_count"`
}

type TotalOrdersByMonth struct {
	Month      int    `json:"month"`
	Bulan      string `json:"bulan"`
	OrderCount int64  `json:"order_count"`
}

type OverviewMonthRevenue struct {
	TotalRevenue int64  `json:"total_revenue"`
	MonthNumber  int    `json:"month_number"`
	Bulan        string `json:"bulan"`
}

type OverviewWeekRevenue struct {
	MonthWeek        string `json:"month_week"`
	TotalTransaction int64  `json:"total_transaction"`
}

type FrequentlyUsedService struct {
	ID           int    `json:"id"`
	ServiceName  string `json:"service_name"`
	ServiceCount int64  `json:"service_count"`
}

type MitraOrderToday struct {
	OrderCount   int64  `json:"order_count"`
	ID           string `json:"id"`
	CompleteName string `json:"complete_name"`
}
