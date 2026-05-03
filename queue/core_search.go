package queue

import (
	"fmt"
	"log"
	"strings"
	"suberes_golang/config"
	"suberes_golang/models"
)

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
	JsonOrderTimes       []models.OrderTransactionRepeat
	GrossAmountCompany   float64
	IsWithTime           bool
	IsCash               bool // true when payment_type == "tunai"; avoids 2 extra DB round-trips
	Limit                int
	Page                 int
}

// MitraResult holds the result of a nearest mitra search.
type MitraResult struct {
	IsAvailableNextTime bool
	PayloadMitra        []models.User
	InitRange           float64
	TriedRange          float64
	// IsAutoBid is true when exactly one mitra was found and that mitra has is_auto_bid = 'yes'.
	// In that case the queue should send ORDER_AUTO_BID instead of ORDER_BROADCAST.
	IsAutoBid      bool
	AutoBidMitraID string
}

// nearestMitraQueryParams is the internal config for buildNearestMitraQuery.
type nearestMitraQueryParams struct {
	Latitude           float64
	Longitude          float64
	MaxRangeKm         float64
	UserGender         string
	IsCash             bool
	IsWithTime         bool
	ServiceDuration    int // in minutes, caller must include the +15 buffer
	GrossAmountCompany float64
	Limit              int // max results; used only for cash+withTime, otherwise 1
}

// GetNearestMitraProduction finds the nearest available mitra using a single
// PostGIS-optimized query instead of an expanding km-loop with multiple DB round-trips.
//
// Original loop behavior is preserved in the ORDER BY of the generated query:
//  1. Smaller km bucket (FLOOR(distance / 1000)) — closer radius first
//  2. is_auto_bid = 'yes' within the same km bucket — auto-bid preferred
//  3. Actual distance for tie-breaking
func GetNearestMitraProduction(params GetNearestMitraProductionParams) (*MitraResult, error) {
	if params.OrderType != "now" {
		return &MitraResult{
			TriedRange:   params.MaxRange,
			PayloadMitra: []models.User{},
		}, nil
	}

	queryStr, queryArgs := buildNearestMitraQuery(nearestMitraQueryParams{
		Latitude:           params.Latitude,
		Longitude:          params.Longitude,
		MaxRangeKm:         params.MaxRange,
		UserGender:         params.UserGender,
		IsCash:             params.IsCash,
		IsWithTime:         params.IsWithTime,
		ServiceDuration:    params.ServiceDuration + 15,
		GrossAmountCompany: params.GrossAmountCompany,
		Limit:              params.Limit,
	})

	var resultQuery []models.User
	if err := config.DB.Raw(queryStr, queryArgs...).Scan(&resultQuery).Error; err != nil {
		log.Printf("[core_search] query error for order subPaymentID=%s: %v", params.SubPaymentID, err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	if len(resultQuery) > 0 {
		isAutoBid := len(resultQuery) == 1 && resultQuery[0].IsAutoBid == "yes"
		autoBidMitraID := ""
		if isAutoBid {
			autoBidMitraID = resultQuery[0].ID
		}
		return &MitraResult{
			IsAvailableNextTime: false,
			PayloadMitra:        resultQuery,
			InitRange:           params.InitialRange,
			IsAutoBid:           isAutoBid,
			AutoBidMitraID:      autoBidMitraID,
		}, nil
	}

	return &MitraResult{
		TriedRange:   params.MaxRange,
		PayloadMitra: []models.User{},
	}, nil
}

// buildNearestMitraQuery builds a single PostGIS-optimized parameterized query that
// replaces the original expanding km-loop + autoBid-toggle loop.
//
// Why this is faster at 100K–500K rows:
//   - ST_DWithin leverages a GiST spatial index → eliminates full table scan
//   - CTEs pre-aggregate debts and order conflicts once instead of per-row correlated subqueries
//   - ORDER BY encodes the same km-bucket + autoBid priority as the original loop
//   - Single DB round-trip instead of up to (maxRange × 2) round-trips
//   - Parameterized query ($n placeholders) prevents SQL injection
//
// Recommended one-time DB migrations for best performance:
//
//	-- 1. Generated geography column so the cast is pre-computed and indexed
//	ALTER TABLE users
//	  ADD COLUMN geom geography(Point, 4326)
//	  GENERATED ALWAYS AS (
//	    ST_SetSRID(ST_MakePoint(longitude::float, latitude::float), 4326)::geography
//	  ) STORED;
//
//	-- 2. GiST index for ST_DWithin spatial lookups
//	CREATE INDEX idx_users_geom ON users USING GIST(geom);
//
//	-- 3. Partial composite index for the static filter predicates
//	CREATE INDEX idx_users_mitra_active
//	  ON users(user_gender, is_auto_bid, account_balance)
//	  WHERE user_type = 'mitra'
//	    AND is_logged_in = '1'
//	    AND is_active = 'yes'
//	    AND is_busy = 'no'
//	    AND is_suspended = '0';
func buildNearestMitraQuery(p nearestMitraQueryParams) (string, []interface{}) {
	var args []interface{}

	// arg registers a value once and returns its $n placeholder.
	// PostgreSQL allows reusing the same $n multiple times in one query,
	// so lon/lat are registered once and referenced wherever needed.
	arg := func(v interface{}) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	// Core spatial args — registered once, reused by $n reference
	lonRef := arg(p.Longitude)               // $1
	latRef := arg(p.Latitude)                // $2
	maxMetersRef := arg(p.MaxRangeKm * 1000) // $3
	genderRef := arg(p.UserGender)           // $4

	// Spatial expression helpers (reference $1/$2, no extra args added).
	//
	// makePointUser pakai kolom `geom` yang sudah pre-computed + di-index GiST
	// (hasil migration SQL di komentar fungsi ini).
	// Kalau migration BELUM dijalankan, ganti "u.geom" kembali ke:
	//   "ST_SetSRID(ST_MakePoint(u.longitude::float, u.latitude::float), 4326)::geography"
	makePointUser := "u.geom"
	makePointCustomer := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography", lonRef, latRef)
	distExpr := fmt.Sprintf("ST_Distance(%s, %s)", makePointUser, makePointCustomer)

	// --- CTEs ---
	var cteList []string

	// Debt CTE: aggregate total hutang per mitra once (replaces per-row correlated subquery)
	// Filter by the same spatial radius so the CTE only processes mitras that could match,
	// instead of scanning the entire tools_credits + tools tables.
	cteList = append(cteList, fmt.Sprintf(`mitra_debts AS (
	SELECT tc.mitra_id, COALESCE(SUM(t.debt_per_week), 0) AS total_hutang
	FROM tools_credits tc
	JOIN tools t ON tc.tool_id = t.id
	WHERE tc.mitra_id IN (
		SELECT u2.id FROM users u2
		WHERE u2.user_type = 'mitra'
		  AND ST_DWithin(u2.geom, %s, %s)
	)
	GROUP BY tc.mitra_id
)`, makePointCustomer, maxMetersRef))

	// Order-time conflict CTEs: only needed when checking schedule window
	var orderJoins, orderFilters string
	if p.IsWithTime {
		// Register minutes once — reused in both CTEs via the same $n
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
)`, minutesRef)) // reuse same $5 — no duplicate arg added

		orderJoins = `
	LEFT JOIN mitra_order_count moc ON moc.mitra_id = u.id
	LEFT JOIN mitra_repeat_order_count mroc ON mroc.mitra_id = u.id`
		orderFilters = `
	AND COALESCE(moc.count_order, 0) = 0
	AND COALESCE(mroc.count_order_repeat, 0) = 0`
	}

	// Cash-specific CTE and filters
	var cashJoins, cashFilters string
	if p.IsCash {
		grossRef := arg(p.GrossAmountCompany) // $5 or $6 depending on IsWithTime
		// NOTE: pending_cash is a global sum of all WAIT_SCHEDULE orders —
		// this matches the original query's intent (system-wide committed cash).
		cteList = append(cteList, `pending_cash AS (
	SELECT COALESCE(SUM(x.gross_amount_company), 0) AS total_pending
	FROM order_transactions x
	WHERE x.order_status = 'WAIT_SCHEDULE'
)`)
		cashJoins = `
	CROSS JOIN pending_cash pc`
		// grossRef reused twice via same $n placeholder — no duplicate arg
		cashFilters = fmt.Sprintf(`
	AND (u.account_balance - pc.total_pending) >= %s
	AND u.account_balance >= %s`, grossRef, grossRef)
	}

	// Only cash+withTime returns multiple mitras (offers sent to a pool).
	// All other variants return the single best match.
	limit := 1
	if p.IsCash && p.IsWithTime {
		if p.Limit > 1 {
			limit = p.Limit
		} else {
			limit = 10
		}
	}

	cteBlock := "WITH " + strings.Join(cteList, ",\n") + "\n"

	// The ORDER BY encodes the same priority as the original loop:
	//   FLOOR(distance_km) → km bucket (closer first)
	//   is_auto_bid = 'yes' → auto-bid preferred within same bucket
	//   distance → tie-break within same bucket + autoBid tier
	query := fmt.Sprintf(`%sSELECT
	u.id, u.firebase_token, u.is_auto_bid, u.account_balance
FROM users u
LEFT JOIN mitra_debts md ON md.mitra_id = u.id%s%s
WHERE
	u.user_type = 'mitra'
	AND u.is_logged_in = '1'
	AND u.is_active = 'yes'
	AND u.is_busy = 'no'
	AND u.is_suspended = '0'
	AND u.user_gender = %s
	AND ST_DWithin(%s, %s, %s)
	AND COALESCE(md.total_hutang, 0) <= u.account_balance%s%s
ORDER BY
	FLOOR(%s / 1000),
	CASE WHEN u.is_auto_bid = 'yes' THEN 0 ELSE 1 END,
	%s
LIMIT %d`,
		cteBlock,
		orderJoins, cashJoins,
		genderRef,
		makePointUser, makePointCustomer, maxMetersRef,
		orderFilters, cashFilters,
		distExpr,
		distExpr,
		limit,
	)

	return query, args
}
