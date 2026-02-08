package helpers

import (
	"fmt"
	"math"

	"github.com/gin-gonic/gin"
)

// RowsDataWrapper: Struktur ini sekarang HANYA berisi docs
type RowsDataWrapper struct {
	Docs interface{} `json:"docs"`
}

// PaginationResponse: Struktur JSON response utama
type PaginationResponse struct {
	CountTotalSize     int             `json:"count_total_size"`     // Jumlah data di halaman ini
	CountTotalPage     int             `json:"count_total_page"`     // Total jumlah halaman
	CountTotal         int64           `json:"count_total"`          // Total seluruh record
	PreviousPage       *string         `json:"previous_page"`        // URL halaman sebelumnya
	PreviousPageNumber *int            `json:"previous_page_number"` // Nomor halaman sebelumnya
	NextPage           *string         `json:"next_page"`            // URL halaman berikutnya
	NextPageNumber     *int            `json:"next_page_number"`     // Nomor halaman berikutnya
	IsLast             string          `json:"is_last"`              // '0' atau '1'
	RowsData           RowsDataWrapper `json:"rows_data"`            // Object pembungkus docs saja
}

// GetPaginationData mengonstruksi response pagination
// Parameter:
// - ctx: Gin Context
// - docs: Slice data (misal []models.Banner)
// - currentSize: Jumlah data di docs (len(docs))
// - page: Halaman saat ini
// - limit: Limit per halaman
// - total: Total record di database
func GetPaginationData(ctx *gin.Context, docs interface{}, currentSize int, page, limit int, total int64) PaginationResponse {
	// 1. Hitung Total Pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	basePath := ctx.Request.URL.Path

	// 2. Logic Previous Page
	var prevPageUrl *string
	var prevPageNum *int

	// Logic: page > 0 && page <= 1 ? null : ... (Sesuai nodejs logic anda)
	// Simplified: Jika page > 1 dan page <= totalPages
	if page > 1 && page <= totalPages+1 { // +1 untuk handle case jika page melebihi total
		pNum := page - 1
		// Pastikan pNum tidak melebihi totalPages jika logic anda mengharuskannya,
		// tapi biasanya prev page selalu page-1 selama page > 1
		if pNum <= totalPages {
			prevPageNum = &pNum
			url := fmt.Sprintf("%s?limit=%d&page=%d", basePath, limit, pNum)
			prevPageUrl = &url
		}
	}

	// 3. Logic Next Page
	var nextPageUrl *string
	var nextPageNum *int

	if page < totalPages {
		nNum := page + 1
		nextPageNum = &nNum
		url := fmt.Sprintf("%s?limit=%d&page=%d", basePath, limit, nNum)
		nextPageUrl = &url
	}

	// 4. Logic Is Last (prev != null && next == null)
	isLast := "0"
	if prevPageUrl != nil && nextPageUrl == nil {
		isLast = "1"
	}

	// 5. Return Response
	return PaginationResponse{
		CountTotalSize:     currentSize,
		CountTotalPage:     totalPages,
		CountTotal:         total,
		PreviousPage:       prevPageUrl,
		PreviousPageNumber: prevPageNum,
		NextPage:           nextPageUrl,
		NextPageNumber:     nextPageNum,
		IsLast:             isLast,
		RowsData: RowsDataWrapper{
			Docs: docs, // Hanya Docs yang dimasukkan
		},
	}
}
