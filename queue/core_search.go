package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"suberes_golang/config"
)

// maxSearchRadiusMeters adalah radius pencarian mitra yang digunakan oleh algoritma scoring.
// Seluruh mitra dalam radius ini di-load ke memori lalu di-rank menggunakan skor komposit.
const maxSearchRadiusMeters = 8000.0

// candidateLimit adalah jumlah maksimum mitra yang diambil dari DB sebelum di-scoring.
const candidateLimit = 50

// GetNearestMitraProductionParams holds all parameters for the mitra search.
type GetNearestMitraProductionParams struct {
	CustomerID           string
	Latitude             float64
	Longitude            float64
	InitialRange         float64
	MaxRange             float64
	UserGender           string
	OrderType            string
	SubPaymentID         int
	IsAutoBid            string
	ServiceDuration      int
	CustomerTimezoneCode string
	CustomerTimeOrder    string
	GrossAmountCompany   float64
	IsWithTime           bool
	IsCash               bool // true when payment_type == "tunai"
	Limit                int
	Page                 int
}

// MitraCandidate holds per-mitra data needed for the scoring algorithm.
type MitraCandidate struct {
	ID                string
	FirebaseToken     *string
	IsAutoBid         string
	AccountBalance    int64
	TotalOrder        int
	TodayOrderCount   int // count of today's orders from order_transactions
	RejectionCount    int
	ResponseMitraRate float64 // cumulative response rate stored in user model (0–1)
	DistanceMeters    float64 // computed by PostGIS
	RatingCount       int     // count of rated orders (is_rated='1') from order_transactions
	AvgRating         float64 // average of rated field from those orders
	Score             float64 // final composite score (higher = better)
}

// MitraResult holds the result of a nearest mitra search.
type MitraResult struct {
	IsAvailableNextTime bool
	// ScoredCandidates is the ranked list (highest score first).
	// Queue handlers store IDs in order_transaction.candidate_mitra_ids for the estafet model.
	ScoredCandidates []MitraCandidate
	InitRange        float64
	TriedRange       float64
	// IsAutoBid is true when the top-ranked mitra has is_auto_bid = 'yes'.
	IsAutoBid      bool
	AutoBidMitraID string
}

// nearestMitraQueryParams is the internal config for buildScoringQuery.
type nearestMitraQueryParams struct {
	Latitude           float64
	Longitude          float64
	UserGender         string
	IsCash             bool
	IsWithTime         bool
	ServiceDuration    int // in minutes, caller must include the +15 buffer
	GrossAmountCompany float64
}

// GetNearestMitraProduction mencari semua mitra yang tersedia dalam radius 8 km,
// menghitung skor komposit untuk setiap mitra, dan mengembalikan daftar kandidat
// yang sudah diurutkan dari skor tertinggi ke terendah.
//
// Algoritma dua tier:
//
//	Tier 1 (total_order < 100):  50% jarak · 25% response rate · 15% cancel rate · 10% fairness
//	Tier 2 (total_order >= 100): 35% jarak · 25% Bayesian rating · 20% response rate · 10% fairness
//
// Daftar kandidat yang diurutkan disimpan di order_transaction.candidate_mitra_ids.
// Queue handler mengirim offer hanya ke kandidat teratas; jika ditolak/timeout 3 menit,
// order diestafetkan ke kandidat berikutnya.
func GetNearestMitraProduction(params GetNearestMitraProductionParams) (*MitraResult, error) {
	log.Printf("[CORE_SEARCH] params | order_type=%s gender=%s lat=%.6f lon=%.6f is_cash=%v is_with_time=%v gross=%.2f",
		params.OrderType, params.UserGender, params.Latitude, params.Longitude,
		params.IsCash, params.IsWithTime, params.GrossAmountCompany)

	if params.OrderType != "now" {
		log.Printf("[CORE_SEARCH] order_type bukan 'now' (%q) → return kosong", params.OrderType)
		return &MitraResult{
			TriedRange:       params.MaxRange,
			ScoredCandidates: []MitraCandidate{},
		}, nil
	}

	queryStr, queryArgs := buildScoringQuery(nearestMitraQueryParams{
		Latitude:           params.Latitude,
		Longitude:          params.Longitude,
		UserGender:         params.UserGender,
		IsCash:             params.IsCash,
		IsWithTime:         params.IsWithTime,
		ServiceDuration:    params.ServiceDuration + 15,
		GrossAmountCompany: params.GrossAmountCompany,
	})

	type rawRow struct {
		ID                string  `gorm:"column:id"`
		FirebaseToken     *string `gorm:"column:firebase_token"`
		IsAutoBid         string  `gorm:"column:is_auto_bid"`
		AccountBalance    int64   `gorm:"column:account_balance"`
		TotalOrder        int     `gorm:"column:total_order"`
		RejectionCount    int     `gorm:"column:rejection_count"`
		ResponseMitraRate float64 `gorm:"column:response_mitra_rate"`
		TodayOrderCount   int     `gorm:"column:today_order_count"`
		RatingCount       int     `gorm:"column:rating_count"`
		AvgRating         float64 `gorm:"column:avg_rating"`
		DistanceMeters    float64 `gorm:"column:distance_meters"`
	}

	var rows []rawRow
	if err := config.DB.Raw(queryStr, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("error executing scoring query: %w", err)
	}

	if len(rows) == 0 {
		return &MitraResult{
			TriedRange:       maxSearchRadiusMeters / 1000,
			ScoredCandidates: []MitraCandidate{},
		}, nil
	}

	// Hitung global average rating (C) untuk Bayesian average pada Tier 2.
	var totalRatingSum float64
	var totalRatingCount int
	for _, r := range rows {
		if r.RatingCount > 0 {
			totalRatingSum += r.AvgRating * float64(r.RatingCount)
			totalRatingCount += r.RatingCount
		}
	}
	globalAvgRating := 4.0 // default ketika belum ada rating
	if totalRatingCount > 0 {
		globalAvgRating = totalRatingSum / float64(totalRatingCount)
	}

	// Bangun dan hitung skor untuk setiap kandidat.
	candidates := make([]MitraCandidate, 0, len(rows))
	for _, r := range rows {
		c := MitraCandidate{
			ID:                r.ID,
			FirebaseToken:     r.FirebaseToken,
			IsAutoBid:         r.IsAutoBid,
			AccountBalance:    r.AccountBalance,
			TotalOrder:        r.TotalOrder,
			TodayOrderCount:   r.TodayOrderCount,
			RejectionCount:    r.RejectionCount,
			ResponseMitraRate: r.ResponseMitraRate,
			DistanceMeters:    r.DistanceMeters,
			RatingCount:       r.RatingCount,
			AvgRating:         r.AvgRating,
		}
		c.Score = computeMitraScore(c, globalAvgRating)
		candidates = append(candidates, c)
	}

	// Urutkan berdasarkan skor tertinggi.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	isAutoBid := false
	autoBidMitraID := ""
	if len(candidates) > 0 && candidates[0].IsAutoBid == "yes" {
		isAutoBid = true
		autoBidMitraID = candidates[0].ID
	}

	logIDs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		logIDs = append(logIDs, fmt.Sprintf("%s(%.3f)", c.ID, c.Score))
	}
	log.Printf("[CORE_SEARCH] %d mitra ditemukan dalam 8km | global_avg_rating=%.2f | ranking: %v",
		len(candidates), globalAvgRating, logIDs)

	return &MitraResult{
		IsAvailableNextTime: false,
		ScoredCandidates:    candidates,
		InitRange:           maxSearchRadiusMeters / 1000,
		TriedRange:          maxSearchRadiusMeters / 1000,
		IsAutoBid:           isAutoBid,
		AutoBidMitraID:      autoBidMitraID,
	}, nil
}

// computeMitraScore menghitung skor komposit untuk satu kandidat mitra.
//
// Tier 1 (total_order < 100) — mitra baru:
//
//	50% distance · 25% response rate · 15% cancel rate · 10% fairness (today orders)
//
// Tier 2 (total_order >= 100) — mitra berpengalaman:
//
//	35% distance · 25% Bayesian avg rating · 20% response rate · 10% fairness
//
// Semua sub-skor berada di kisaran [0,1]; skor komposit adalah weighted sum.
func computeMitraScore(c MitraCandidate, globalAvgRating float64) float64 {
	// ── Distance score (makin dekat, makin tinggi) ────────────────────────────
	// Dinormalisasi ke [0,1] dalam radius 8 km; mitra di 0 m mendapat 1.0.
	distScore := math.Max(0, 1.0-c.DistanceMeters/maxSearchRadiusMeters)

	// ── Response rate ─────────────────────────────────────────────────────────
	// Disimpan sebagai 0–1 di user.response_mitra_rate.
	// Mitra baru (belum ada history) diberi nilai netral 0.5 agar tidak dirugikan.
	responseScore := c.ResponseMitraRate
	if c.TotalOrder == 0 {
		responseScore = 0.5
	}

	// ── Cancel / rejection score ──────────────────────────────────────────────
	// Makin sedikit rejection → makin tinggi skor. Cap di 50 rejection → 0.
	cancelScore := math.Max(0, 1.0-float64(c.RejectionCount)/50.0)

	// ── Fairness score (jumlah orderan hari ini) ──────────────────────────────
	// Mitra dengan lebih sedikit order hari ini mendapat prioritas lebih tinggi.
	// Hyperbolic decay: 0 order → 1.0, 1 order → 0.5, 2 order → 0.33, dst.
	fairnessScore := 1.0 / (1.0 + float64(c.TodayOrderCount))

	if c.TotalOrder < 100 {
		// Tier 1 – mitra baru
		// 50% jarak · 25% response rate · 15% cancel rate · 10% fairness
		return 0.50*distScore +
			0.25*responseScore +
			0.15*cancelScore +
			0.10*fairnessScore
	}

	// ── Bayesian average rating ───────────────────────────────────────────────
	// Formula: (v×R + m×C) / (v + m)
	//   v = jumlah orderan yang diberi rating untuk mitra ini
	//   R = rata-rata rating mitra ini
	//   m = 50 (minimum vote agar rating dianggap valid/terpercaya)
	//   C = rata-rata rating global dari semua kandidat
	const bayesianM = 50.0
	v := float64(c.RatingCount)
	R := c.AvgRating
	bayesian := (v*R + bayesianM*globalAvgRating) / (v + bayesianM)
	// Normalisasi dari skala [1,5] → [0,1]
	ratingScore := math.Max(0, math.Min(1, (bayesian-1.0)/4.0))

	// Tier 2 – mitra berpengalaman
	// 35% jarak · 25% Bayesian rating · 20% response rate · 10% fairness
	return 0.35*distScore +
		0.25*ratingScore +
		0.20*responseScore +
		0.10*fairnessScore
}

// ExtractCandidateIDs mengembalikan daftar mitra ID yang sudah diurutkan berdasarkan skor.
func ExtractCandidateIDs(result *MitraResult) []string {
	ids := make([]string, 0, len(result.ScoredCandidates))
	for _, c := range result.ScoredCandidates {
		ids = append(ids, c.ID)
	}
	return ids
}

// MarshalCandidateIDs meng-encode slice mitra IDs ke JSON untuk disimpan di DB.
func MarshalCandidateIDs(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(ids)
	return string(b)
}

// UnmarshalCandidateIDs men-decode JSON array mitra IDs dari order_transaction.
func UnmarshalCandidateIDs(raw string) []string {
	if raw == "" || raw == "[]" {
		return nil
	}
	var ids []string
	_ = json.Unmarshal([]byte(raw), &ids)
	return ids
}

// buildScoringQuery membangun query PostGIS yang mengambil semua kolom yang diperlukan
// untuk algoritma scoring dua-tier dalam satu DB round-trip.
//
// Query ini menggantikan buildNearestMitraQuery yang lama. Perbedaan utama:
//   - Mengambil data rating, today_order, rejection_count, response_mitra_rate per mitra
//   - Radius pencarian tetap 8 km (tidak lagi adaptive)
//   - Tidak ada ORDER BY di sisi DB — pengurutan dilakukan di Go setelah scoring
//   - LIMIT candidateLimit (50) untuk membatasi jumlah kandidat yang diproses
//
// Recommended one-time DB migrations (sama seperti sebelumnya):
//
//	ALTER TABLE users ADD COLUMN geom geography(Point, 4326)
//	  GENERATED ALWAYS AS (
//	    ST_SetSRID(ST_MakePoint(longitude::float, latitude::float), 4326)::geography
//	  ) STORED;
//	CREATE INDEX idx_users_geom ON users USING GIST(geom);
func buildScoringQuery(p nearestMitraQueryParams) (string, []interface{}) {
	var args []interface{}

	// arg mendaftarkan satu nilai dan mengembalikan placeholder $n-nya.
	arg := func(v interface{}) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	lonRef := arg(p.Longitude)                 // $1
	latRef := arg(p.Latitude)                  // $2
	maxMetersRef := arg(maxSearchRadiusMeters) // $3  (8000 m)
	genderRef := arg(p.UserGender)             // $4

	// makePointUser pakai COALESCE: utamakan kolom `geom` pre-computed (GiST index) untuk
	// performa, tapi fallback ke ST_MakePoint jika geom masih NULL (mitra lama sebelum migration).
	makePointUser := "COALESCE(u.geom, ST_SetSRID(ST_MakePoint(u.longitude::float, u.latitude::float), 4326)::geography)"
	makePointCustomer := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography", lonRef, latRef)
	distExpr := fmt.Sprintf("ST_Distance(%s, %s)", makePointUser, makePointCustomer)

	var cteList []string

	// ── Debt CTE (sama seperti sebelumnya) ───────────────────────────────────
	cteList = append(cteList, fmt.Sprintf(`mitra_debts AS (
	SELECT tc.mitra_id, COALESCE(SUM(t.debt_per_week), 0) AS total_hutang
	FROM tools_credits tc
	JOIN tools t ON tc.tool_id = t.id
	WHERE tc.mitra_id IN (
		SELECT u2.id FROM users u2
		WHERE u2.user_type = 'mitra'
		  AND ST_DWithin(COALESCE(u2.geom, ST_SetSRID(ST_MakePoint(u2.longitude::float, u2.latitude::float), 4326)::geography), %s, %s)
	)
	GROUP BY tc.mitra_id
)`, makePointCustomer, maxMetersRef))

	// ── Jumlah order hari ini per mitra (untuk skor fairness) ─────────────────
	cteList = append(cteList, `mitra_today_orders AS (
	SELECT mitra_id, COUNT(*) AS today_count
	FROM order_transactions
	WHERE mitra_id IS NOT NULL
	  AND order_status IN ('FINISH','OTW','ON_PROGRESS','WAIT_SCHEDULE')
	  AND created_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')
	GROUP BY mitra_id
)`)

	// ── Agregat rating per mitra (untuk Bayesian average di Tier 2) ───────────
	cteList = append(cteList, `mitra_ratings AS (
	SELECT mitra_id,
	       COUNT(*) AS rating_count,
	       COALESCE(AVG(rated::float), 0) AS avg_rating
	FROM order_transactions
	WHERE mitra_id IS NOT NULL
	  AND is_rated = '1'
	  AND rated > 0
	GROUP BY mitra_id
)`)

	// ── CTE konflik waktu (hanya saat IsWithTime = true) ─────────────────────
	var orderJoins, orderFilters string
	if p.IsWithTime {
		minutesRef := arg(p.ServiceDuration) // $5
		cteList = append(cteList, fmt.Sprintf(`mitra_order_count AS (
	SELECT ot.mitra_id,
		SUM(CASE
			WHEN EXTRACT(EPOCH FROM (ot.order_time - NOW())) / 60 > 0
			 AND EXTRACT(EPOCH FROM (ot.order_time - NOW())) / 60 < %s
			THEN 1 ELSE 0 END
		) AS count_order
	FROM order_transactions ot
	WHERE ot.order_status = 'WAIT_SCHEDULE'
	GROUP BY ot.mitra_id
)`, minutesRef))
		cteList = append(cteList, fmt.Sprintf(`mitra_repeat_order_count AS (
	SELECT otr.mitra_id,
		SUM(CASE
			WHEN EXTRACT(EPOCH FROM (otr.order_time - NOW())) / 60 > 0
			 AND EXTRACT(EPOCH FROM (otr.order_time - NOW())) / 60 < %s
			THEN 1 ELSE 0 END
		) AS count_order_repeat
	FROM order_transaction_repeats otr
	GROUP BY otr.mitra_id
)`, minutesRef))
		orderJoins = `
	LEFT JOIN mitra_order_count moc ON moc.mitra_id = u.id
	LEFT JOIN mitra_repeat_order_count mroc ON mroc.mitra_id = u.id`
		orderFilters = `
	AND COALESCE(moc.count_order, 0) = 0
	AND COALESCE(mroc.count_order_repeat, 0) = 0`
	}

	// ── CTE khusus tunai (cek saldo mitra) ───────────────────────────────────
	// pending_cash dikelompokkan PER MITRA agar deduks saldo hanya memperhitungkan
	// komitmen mitra tersebut, bukan seluruh sistem (bug lama: CROSS JOIN sistem-wide).
	var cashJoins, cashFilters string
	if p.IsCash {
		grossRef := arg(p.GrossAmountCompany)
		cteList = append(cteList, `pending_cash AS (
	SELECT mitra_id, COALESCE(SUM(x.gross_amount_company), 0) AS total_pending
	FROM order_transactions x
	WHERE x.order_status = 'WAIT_SCHEDULE'
	  AND x.mitra_id IS NOT NULL
	GROUP BY mitra_id
)`)
		cashJoins = `
	LEFT JOIN pending_cash pc ON pc.mitra_id = u.id`
		cashFilters = fmt.Sprintf(`
	AND (u.account_balance - COALESCE(pc.total_pending, 0)) >= %s
	AND u.account_balance >= %s`, grossRef, grossRef)
	}

	cteBlock := "WITH " + strings.Join(cteList, ",\n") + "\n"

	// Query mengambil semua kolom yang dibutuhkan untuk scoring di Go-side.
	// ORDER BY sengaja dihilangkan — pengurutan dilakukan setelah scoring.
	query := fmt.Sprintf(`%sSELECT
	u.id,
	u.firebase_token,
	u.is_auto_bid,
	u.account_balance,
	u.total_order,
	u.rejection_count,
	COALESCE(u.response_mitra_rate, 0) AS response_mitra_rate,
	COALESCE(mto.today_count, 0)       AS today_order_count,
	COALESCE(mr.rating_count, 0)       AS rating_count,
	COALESCE(mr.avg_rating, 0)         AS avg_rating,
	%s                                  AS distance_meters
FROM users u
LEFT JOIN mitra_debts md          ON md.mitra_id  = u.id
LEFT JOIN mitra_today_orders mto  ON mto.mitra_id = u.id
LEFT JOIN mitra_ratings mr        ON mr.mitra_id  = u.id%s%s
WHERE
	u.user_type    = 'mitra'
	AND u.is_logged_in = '1'
	AND u.is_active    = 'yes'
	AND u.is_busy      = 'no'
	AND u.is_suspended = '0'
	AND u.user_gender  = %s
	AND ST_DWithin(%s, %s, %s)
	AND COALESCE(md.total_hutang, 0) <= u.account_balance%s%s
LIMIT %d`,
		cteBlock,
		distExpr,
		orderJoins, cashJoins,
		genderRef,
		makePointUser, makePointCustomer, maxMetersRef,
		orderFilters, cashFilters,
		candidateLimit,
	)

	return query, args
}
