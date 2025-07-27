package dto

import (
	"fmt"
	"math"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Total        int    `json:"total"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
	TotalPages   int    `json:"total_pages"`
	PreviousLink string `json:"previous_page"`
	NextLink     string `json:"next_page"`
}

func NewPagination(ctx *gin.Context, total, page, limit int) Pagination {
	offset := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	baseURL := ctx.Request.URL.Path
	query := ctx.Request.URL.Query()

	cloneQuery := func(q url.Values) url.Values {
		clone := url.Values{}
		for k, v := range q {
			clone[k] = append([]string{}, v...)
		}
		return clone
	}

	var prevPage, nextPage string

	if page > 1 {
		q := cloneQuery(query)
		q.Set("page", strconv.Itoa(page-1))
		q.Set("limit", strconv.Itoa(limit))
		prevPage = fmt.Sprintf("%s?%s", baseURL, q.Encode())
	}

	if page < totalPages {
		q := cloneQuery(query)
		q.Set("page", strconv.Itoa(page+1))
		q.Set("limit", strconv.Itoa(limit))
		nextPage = fmt.Sprintf("%s?%s", baseURL, q.Encode())
	}

	return Pagination{
		Total:        total,
		Page:         page,
		Limit:        limit,
		Offset:       offset,
		TotalPages:   totalPages,
		PreviousLink: prevPage,
		NextLink:     nextPage,
	}
}
