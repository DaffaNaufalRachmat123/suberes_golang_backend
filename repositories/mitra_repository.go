package repositories

import (
	"fmt"
	"strings"
	"suberes_golang/models"
	"time"

	"gorm.io/gorm"
)

type MitraRepository struct {
	DB *gorm.DB
}

type MitraSearchQuery struct {
	Latitude             float64
	Longitude            float64
	IsActive             string
	IsBusy               string
	UserRating1          float64
	UserRating2          float64
	Distance             int
	IsAutoBid            string
	MinutesSubServices   int
	UserGender           string
	GrossAmountCompany   float64
	Page                 int
	Limit                int
	CustomerID           string
	SubPaymentID         int
	OrderType            string
	ServiceDuration      int
	CustomerTimezoneCode string
	CustomerTimeOrder    string
	JsonOrderTimes       string
	IsWithTime           bool
	InitialRange         int
	MaxRange             int
}

type MitraSearchResult struct {
	CountOrder       int     `gorm:"column:count_order"`
	CountOrderRepeat int     `gorm:"column:count_order_repeat"`
	ID               string  `gorm:"column:id"`
	FirebaseToken    string  `gorm:"column:firebase_token"`
	IsAutoBid        string  `gorm:"column:is_auto_bid"`
	AccountBalance   float64 `gorm:"column:account_balance"`
	TotalHutang      float64 `gorm:"column:total_hutang"`
	Distance         float64 `gorm:"column:distance"`
	MoneyLeft        float64 `gorm:"column:money_left"`
}

type TodayOrderCount struct {
	OrderCount        int `gorm:"column:order_count"`
	PendapatanHariIni int `gorm:"column:pendapatan_hari_ini"`
}

func (r *MitraRepository) FindMitraByID(id string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Where("id = ? AND user_type = ?", id, "mitra").First(&mitra).Error
	return &mitra, err
}

func (r *MitraRepository) FindMitraByEmail(email string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Where("email = ? AND user_type = ?", email, "mitra").First(&mitra).Error
	return &mitra, err
}

func (r *MitraRepository) UpdateMitra(tx *gorm.DB, mitra *models.User) error {
	return tx.Save(mitra).Error
}

func (r *MitraRepository) IncrementRejectionCount(tx *gorm.DB, mitraID string) error {
	return tx.Table("users").
		Where("id = ? AND user_type = ?", mitraID, "mitra").
		UpdateColumn("rejection_count", gorm.Expr("rejection_count + 1")).Error
}

func (r *MitraRepository) CreateMitra(tx *gorm.DB, mitra *models.User) error {
	return tx.Create(mitra).Error
}

func (r *MitraRepository) IsRunningOrder(mitraID string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Where("id = ? AND user_type = ? AND is_busy = ?", mitraID, "mitra", "yes").First(&mitra).Error
	return &mitra, err
}

func (r *MitraRepository) GetMitraProfile(id string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Select("id, user_profile_image, complete_name, email, is_mitra_activated, user_level, color_code_level, country_code, phone_number, user_rating, user_gender, place_of_birth, EXTRACT(YEAR FROM AGE(date_of_birth::date))::text as age").Where("id = ? AND user_type = ?", id, "mitra").First(&mitra).Error
	return &mitra, err
}

func (r *MitraRepository) GetTodayOrderCount(mitraID string, startTime time.Time, endTime time.Time) (*TodayOrderCount, error) {
	var result TodayOrderCount
	err := r.DB.Model(&models.OrderTransaction{}).Select("SUM(gross_amount_mitra) as pendapatan_hari_ini, COUNT(id) as order_count").Where("mitra_id = ? AND order_status = 'FINISH' AND order_time BETWEEN ? AND ?", mitraID, startTime, endTime).Scan(&result).Error
	return &result, err
}

func (r *MitraRepository) GetTodayOrderRepeatCount(mitraID string, startTime time.Time, endTime time.Time) (*TodayOrderCount, error) {
	var result TodayOrderCount
	err := r.DB.Model(&models.OrderTransactionRepeat{}).Select("SUM(gross_amount_mitra) as pendapatan_hari_ini, COUNT(id) as order_count").Where("mitra_id = ? AND order_status = 'FINISH' AND order_time BETWEEN ? AND ?", mitraID, startTime, endTime).Scan(&result).Error
	return &result, err
}

func (r *MitraRepository) GetTotalCicilan(mitraID string) (int, error) {
	var totalHutang int
	err := r.DB.Model(&models.SubToolCredit{}).Select("COALESCE(SUM(amount_paid), 0)").Where("mitra_id = ? AND paid_status = '0'", mitraID).Row().Scan(&totalHutang)
	return totalHutang, err
}

// GetMitraShowPhone returns only id, country_code, and phone_number for a mitra.
func (r *MitraRepository) GetMitraShowPhone(mitraID string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Select("id, country_code, phone_number").
		Where("id = ? AND user_type = ?", mitraID, "mitra").
		First(&mitra).Error
	return &mitra, err
}

type MitraSaldoData struct {
	AccountBalance int64 `json:"account_balance"`
	Result         int64 `json:"result"`
}

type OrderDataItem struct {
	GrossAmountMitra string `json:"gross_amount_mitra"`
	DayOfWeek        string `json:"day_of_week"`
}

// GetMitraSaldo returns account_balance for a mitra.
func (r *MitraRepository) GetMitraSaldo(mitraID string) (*MitraSaldoData, error) {
	var data MitraSaldoData
	err := r.DB.Table("users").
		Select("account_balance, account_balance as result").
		Where("id = ? AND user_type = ?", mitraID, "mitra").
		Take(&data).Error
	return &data, err
}

// GetMitraOrderDataSaldo returns gross_amount_mitra summed per day-of-week for a mitra.
func (r *MitraRepository) GetMitraOrderDataSaldo(mitraID string) ([]OrderDataItem, error) {
	var items []OrderDataItem
	err := r.DB.Table("order_transactions").
		Select(`
			CAST(COALESCE(SUM(gross_amount_mitra), 0) AS TEXT) AS gross_amount_mitra,
			CASE EXTRACT(DOW FROM created_at)
				WHEN 0 THEN 'Minggu'
				WHEN 1 THEN 'Senin'
				WHEN 2 THEN 'Selasa'
				WHEN 3 THEN 'Rabu'
				WHEN 4 THEN 'Kamis'
				WHEN 5 THEN 'Jumat'
				WHEN 6 THEN 'Sabtu'
				ELSE '-'
			END AS day_of_week
		`).
		Where("mitra_id = ? AND order_status = ?", mitraID, "FINISH").
		Group("EXTRACT(DOW FROM created_at)").
		Order("EXTRACT(DOW FROM created_at)").
		Scan(&items).Error
	if items == nil {
		items = []OrderDataItem{}
	}
	return items, err
}

// GetMitraSaldoProfile returns basic mitra profile fields for the saldo page.
func (r *MitraRepository) GetMitraSaldoProfile(mitraID string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Select("id, complete_name, user_profile_image, country_code, phone_number, email").
		Where("id = ? AND user_type = ?", mitraID, "mitra").
		First(&mitra).Error
	return &mitra, err
}

func (r *MitraRepository) buildSearchQueryNowCash(query MitraSearchQuery) *gorm.DB {
	var searchString strings.Builder
	searchString.WriteString("SELECT ")
	searchString.WriteString(fmt.Sprintf("COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), 'UTC', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), 'UTC', b.timezone_code), b.timezone_code, 'UTC'), b.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b WHERE b.mitra_id = a.id AND b.order_status = 'WAIT_SCHEDULE'),0) AS count_order, ", query.MinutesSubServices))
	searchString.WriteString(fmt.Sprintf("COALESCE((SELECT SUM(CASE WHEN TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), 'UTC', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) > 0 AND TIMESTAMPDIFF(MINUTE, CONVERT_TZ(CONVERT_TZ(NOW(), 'UTC', b.timezone_code), b.timezone_code, 'UTC'), c.order_time) < %d THEN 1 ELSE 0 END) FROM order_transactions b LEFT JOIN order_transaction_repeats c ON c.order_id = b.id WHERE c.order_id = b.id AND c.mitra_id = a.id),0) AS count_order_repeat, ", query.MinutesSubServices))
	searchString.WriteString("a.id, a.firebase_token, a.is_auto_bid, a.account_balance, ")
	searchString.WriteString("(SELECT COALESCE(SUM(c.debt_per_week),0) FROM tools_credits b LEFT JOIN tools c ON b.tool_id = c.id WHERE b.mitra_id = a.id) AS total_hutang, ")
	searchString.WriteString(fmt.Sprintf("(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) AS distance, ", query.Latitude, query.Longitude, query.Latitude))
	searchString.WriteString("(a.account_balance - COALESCE((SELECT SUM(x.gross_amount_company) FROM order_transactions x WHERE x.order_status = 'WAIT_SCHEDULE'),0)) AS money_left ")
	searchString.WriteString("FROM users a ")
	searchString.WriteString(fmt.Sprintf("WHERE a.user_gender = '%s' AND a.is_logged_in = '1' AND a.is_active = '%s' AND a.is_busy = '%s' AND a.is_auto_bid = '%s' AND a.is_suspended = '0' AND a.user_type = 'mitra' ", query.UserGender, query.IsActive, query.IsBusy, query.IsAutoBid))
	searchString.WriteString(fmt.Sprintf("HAVING total_hutang <= a.account_balance AND money_left >= %f AND distance <= %d AND a.account_balance >= %f AND count_order = 0 AND count_order_repeat = 0", query.GrossAmountCompany, query.Distance, query.GrossAmountCompany))

	return r.DB.Raw(searchString.String())
}

func (r *MitraRepository) GetNearestMitra(query MitraSearchQuery) ([]MitraSearchResult, error) {
	var results []MitraSearchResult

	tx := r.buildSearchQueryNowCash(query)

	if err := tx.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MitraRepository) FindMitraForInvite(id string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Where("id = ? AND user_type = ? AND is_mitra_activated = ?", id, "mitra", "0").First(&mitra).Error
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

func (r *MitraRepository) UpdateMitraInvited(tx *gorm.DB, id string) error {
	return tx.Table("users").Where("id = ? AND user_type = ? AND is_mitra_invited = ?", id, "mitra", "0").Update("is_mitra_invited", "1").Error
}

// UpdateMitraTrainingStatus updates is_mitra_accepted or is_mitra_invited based on status (successful|failed).
// Target: users where id=id AND user_type='mitra' AND is_mitra_invited='1' AND is_mitra_activated='0'.
func (r *MitraRepository) UpdateMitraTrainingStatus(tx *gorm.DB, id, status string) error {
	base := tx.Table("users").Where("id = ? AND user_type = ? AND is_mitra_invited = ? AND is_mitra_activated = ?", id, "mitra", "1", "0")
	switch status {
	case "successful":
		return base.Update("is_mitra_accepted", "1").Error
	case "failed":
		return base.Update("is_mitra_invited", "0").Error
	}
	return nil
}

func (r *MitraRepository) FindMitraForActivation(id string) (*models.User, error) {
	var mitra models.User
	err := r.DB.Where("id = ? AND user_type = ? AND is_mitra_activated = ?", id, "mitra", "0").First(&mitra).Error
	if err != nil {
		return nil, err
	}
	return &mitra, nil
}

func (r *MitraRepository) UpdateMitraActivationPayload(tx *gorm.DB, id string, payload map[string]interface{}) error {
	return tx.Table("users").Where("id = ? AND user_type = ? AND is_mitra_activated = ?", id, "mitra", "0").Updates(payload).Error
}

// UpdateMitraByID updates mitra fields by id only, with no re-fetch.
func (r *MitraRepository) UpdateMitraByID(tx *gorm.DB, id string, payload map[string]interface{}) error {
	return tx.Table("users").Where("id = ? AND user_type = ?", id, "mitra").Updates(payload).Error
}

func NewMitraRepository(db *gorm.DB) *MitraRepository {
	return &MitraRepository{DB: db}
}
