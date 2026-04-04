package services

import (
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"gorm.io/gorm"
)

// ── Service ───────────────────────────────────────────────────────────────────

type OrderHistoryService struct {
	DB                          *gorm.DB
	OrderTransactionRepo        *repositories.OrderTransactionRepository
	OrderTransactionRepeatsRepo *repositories.OrderTransactionRepeatsRepository
}

// ── Computed-field result types ───────────────────────────────────────────────

// ComingSoonWithCountdown wraps an order with a client-facing countdown in seconds.
type ComingSoonWithCountdown struct {
	models.OrderTransaction
	CountDownOrderActive int64 `json:"count_down_order_active"`
}

// RepeatOrderItem wraps an order with global repeat totals appended per-row.
type RepeatOrderItem struct {
	models.OrderTransaction
	TotalOrder         int64 `json:"total_order"`
	TotalFinishedCount int64 `json:"total_finished_count"`
}

// RepeatListItem wraps a sub-order with timeout counters in seconds.
type RepeatListItem struct {
	models.OrderTransactionRepeat
	TimeoutRepeatFirst int64 `json:"timeout_repeat_first"`
	TimeoutRepeat      int64 `json:"timeout_repeat"`
}

// Count response types — match the original JS field names.

type MitraCanceledCountResp struct {
	MitraID            string `json:"mitra_id"`
	OrderCanceledTotal int64  `json:"order_canceled_total"`
}

type MitraDoneCountResp struct {
	MitraID        string `json:"mitra_id"`
	OrderDoneTotal int64  `json:"order_done_total"`
}

type MitraComingSoonCountResp struct {
	MitraID              string `json:"mitra_id"`
	OrderComingSoonTotal int64  `json:"order_coming_soon_total"`
}

type MitraRepeatCountResp struct {
	MitraID          string `json:"mitra_id"`
	OrderRepeatTotal int64  `json:"order_repeat_total"`
}

type CustomerPendingCountResp struct {
	CustomerID        string `json:"customer_id"`
	OrderPendingTotal int64  `json:"order_pending_total"`
}

// ── Date-range helpers ────────────────────────────────────────────────────────

// getStartEndDate converts "YYYY-MM-DD" → full-day range as string timestamps.
func getStartEndDate(dateStr string) (start, end string) {
	return dateStr + " 00:00:00", dateStr + " 23:59:59"
}

// parseDateRangeFromQuery handles the 4 cases from the original JS helpers:
//   - only startDate → full day of startDate
//   - only endDate   → full day of endDate
//   - both           → start of startDate through end of endDate
//   - neither        → empty strings (no date filter applied)
func parseDateRangeFromQuery(startDate, endDate string) (start, end string) {
	switch {
	case startDate != "" && endDate == "":
		return getStartEndDate(startDate)
	case startDate == "" && endDate != "":
		return getStartEndDate(endDate)
	case startDate != "" && endDate != "":
		s, _ := getStartEndDate(startDate)
		_, e := getStartEndDate(endDate)
		return s, e
	default:
		return "", ""
	}
}

// ── Order Canceleds ───────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetCanceledDatesByMitra(mitraID string) ([]repositories.OrderHistoryDateResult, error) {
	return s.OrderTransactionRepo.FindCanceledDatesByMitra(mitraID)
}

func (s *OrderHistoryService) GetCanceledCountByMitra(mitraID string) (*MitraCanceledCountResp, error) {
	total, err := s.OrderTransactionRepo.CountCanceledByMitra(mitraID)
	if err != nil {
		return nil, err
	}
	return &MitraCanceledCountResp{MitraID: mitraID, OrderCanceledTotal: total}, nil
}

func (s *OrderHistoryService) GetCanceledByMitraAndDate(mitraID, orderTime, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	start, end := getStartEndDate(orderTime)
	return s.OrderTransactionRepo.FindCanceledByMitraAndDate(mitraID, start, end, search, page, limit)
}

func (s *OrderHistoryService) GetCanceledForCustomer(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	start, end := parseDateRangeFromQuery(startDate, endDate)
	return s.OrderTransactionRepo.FindCanceledForCustomerPagedFull(customerID, start, end, page, limit)
}

// ── Order Dones ───────────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetDoneDatesByMitra(mitraID string) ([]repositories.OrderHistoryDateResult, error) {
	return s.OrderTransactionRepo.FindDoneDatesByMitra(mitraID)
}

func (s *OrderHistoryService) GetDoneCountByMitra(mitraID string) (*MitraDoneCountResp, error) {
	total, err := s.OrderTransactionRepo.CountDoneByMitra(mitraID)
	if err != nil {
		return nil, err
	}
	return &MitraDoneCountResp{MitraID: mitraID, OrderDoneTotal: total}, nil
}

func (s *OrderHistoryService) GetDoneByMitraAndDate(mitraID, orderTime, search string, page, limit int) ([]models.OrderTransaction, int64, error) {
	start, end := getStartEndDate(orderTime)
	return s.OrderTransactionRepo.FindDoneByMitraAndDate(mitraID, start, end, search, page, limit)
}

func (s *OrderHistoryService) GetDoneForCustomer(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	start, end := parseDateRangeFromQuery(startDate, endDate)
	return s.OrderTransactionRepo.FindDoneForCustomerPagedFull(customerID, start, end, page, limit)
}

// ── Order Coming Soon ─────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetComingSoonDatesByMitra(mitraID string) ([]repositories.OrderHistoryDateResult, error) {
	return s.OrderTransactionRepo.FindComingSoonDatesByMitra(mitraID)
}

func (s *OrderHistoryService) GetComingSoonCountByMitra(mitraID string) (*MitraComingSoonCountResp, error) {
	total, err := s.OrderTransactionRepo.CountComingSoonByMitra(mitraID)
	if err != nil {
		return nil, err
	}
	return &MitraComingSoonCountResp{MitraID: mitraID, OrderComingSoonTotal: total}, nil
}

// GetComingSoonByMitraAndDate returns orders with a per-row countdown in seconds.
func (s *OrderHistoryService) GetComingSoonByMitraAndDate(mitraID, orderTime, search string, page, limit int) ([]ComingSoonWithCountdown, int64, error) {
	start, end := getStartEndDate(orderTime)
	orders, total, err := s.OrderTransactionRepo.FindComingSoonByMitraAndDate(mitraID, start, end, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	items := make([]ComingSoonWithCountdown, len(orders))
	for i, o := range orders {
		items[i] = ComingSoonWithCountdown{
			OrderTransaction:     o,
			CountDownOrderActive: int64(o.OrderTime.Sub(now).Seconds()),
		}
	}
	return items, total, nil
}

func (s *OrderHistoryService) GetComingSoonForCustomer(customerID, startDate, endDate string, page, limit int) ([]models.OrderTransaction, int64, error) {
	start, end := parseDateRangeFromQuery(startDate, endDate)
	return s.OrderTransactionRepo.FindComingSoonForCustomerPagedFull(customerID, start, end, page, limit)
}

// ── Order Repeat ──────────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetRepeatDatesByMitra(mitraID string) ([]repositories.OrderHistoryDateResult, error) {
	return s.OrderTransactionRepo.FindRepeatDatesByMitra(mitraID)
}

func (s *OrderHistoryService) GetRepeatCountByMitra(mitraID string) (*MitraRepeatCountResp, error) {
	total, err := s.OrderTransactionRepo.CountRepeatByMitra(mitraID)
	if err != nil {
		return nil, err
	}
	return &MitraRepeatCountResp{MitraID: mitraID, OrderRepeatTotal: total}, nil
}

// GetRepeatByMitraAndDate returns orders with global mitra repeat totals.
func (s *OrderHistoryService) GetRepeatByMitraAndDate(mitraID, orderTime, search string, page, limit int) ([]RepeatOrderItem, int64, error) {
	start, end := getStartEndDate(orderTime)
	orders, total, err := s.OrderTransactionRepo.FindRepeatByMitraAndDate(mitraID, start, end, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	totalOrder, _ := s.OrderTransactionRepo.CountRepeatOrdersForMitra(mitraID)
	totalFinished, _ := s.OrderTransactionRepo.CountFinishedRepeatForMitra(mitraID)

	items := make([]RepeatOrderItem, len(orders))
	for i, o := range orders {
		items[i] = RepeatOrderItem{
			OrderTransaction:   o,
			TotalOrder:         totalOrder,
			TotalFinishedCount: totalFinished,
		}
	}
	return items, total, nil
}

// GetRepeatForCustomer returns repeat orders with global system-wide repeat totals
// (matches the original JS query which has no customer filter on the sub-counts).
func (s *OrderHistoryService) GetRepeatForCustomer(customerID string, page, limit int) ([]RepeatOrderItem, int64, error) {
	orders, total, err := s.OrderTransactionRepo.FindRepeatForCustomerPagedFull(customerID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	totalOrder, _ := s.OrderTransactionRepo.CountAllRepeatOrders()
	totalFinished, _ := s.OrderTransactionRepo.CountAllFinishedRepeatOrders()

	items := make([]RepeatOrderItem, len(orders))
	for i, o := range orders {
		items[i] = RepeatOrderItem{
			OrderTransaction:   o,
			TotalOrder:         totalOrder,
			TotalFinishedCount: totalFinished,
		}
	}
	return items, total, nil
}

// GetRepeatList returns paginated repeat sub-orders with timeout fields in seconds.
func (s *OrderHistoryService) GetRepeatList(orderID, mitraID, customerID string, page, limit int) ([]RepeatListItem, int64, error) {
	repeats, total, err := s.OrderTransactionRepeatsRepo.FindRepeatListByOrderPaged(orderID, mitraID, customerID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	items := make([]RepeatListItem, len(repeats))
	for i, r := range repeats {
		items[i] = RepeatListItem{
			OrderTransactionRepeat: r,
			TimeoutRepeatFirst:     int64(r.OrderTime.Sub(now).Seconds()),
			TimeoutRepeat:          int64(r.OrderTime.Add(3 * time.Minute).Sub(now).Seconds()),
		}
	}
	return items, total, nil
}

// ── Order Pending ─────────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetPendingCountByCustomer(customerID string) (*CustomerPendingCountResp, error) {
	total, err := s.OrderTransactionRepo.CountPendingByCustomer(customerID)
	if err != nil {
		return nil, err
	}
	return &CustomerPendingCountResp{CustomerID: customerID, OrderPendingTotal: total}, nil
}

func (s *OrderHistoryService) GetPendingByMitra(mitraID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindPendingByMitraPaged(mitraID, page, limit)
}

func (s *OrderHistoryService) GetPendingByCustomer(customerID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindPendingByCustomerPaged(customerID, page, limit)
}

// ── Order Running ─────────────────────────────────────────────────────────────

func (s *OrderHistoryService) GetRunningForCustomer(customerID string, page, limit int) ([]models.OrderTransaction, int64, error) {
	return s.OrderTransactionRepo.FindRunningForCustomerPagedFull(customerID, page, limit)
}
