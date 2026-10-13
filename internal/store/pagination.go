package store

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	// errInvalidLimitValue is returned when the pagination limit query parameter is invalid.
	errInvalidLimitValue = errors.New("invalid limit value")
	// errInvalidOffsetValue is returned when the pagination offset query parameter is invalid.
	errInvalidOffsetValue = errors.New("invalid offset value")
	// errInvalidSortValue is returned when the sort direction query parameter is invalid.
	errInvalidSortValue = errors.New("invalid sort value")
)

const (
	// SortModeASC represents ascending sort order.
	SortModeASC = "asc"
	// SortModeDESC represents descending sort order.
	SortModeDESC = "desc"
)

// PaginatedFeedQuery holds filtering, sorting, and pagination parameters for feed queries.
type PaginatedFeedQuery struct {
	Limit  int      `json:"limit"  validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort"   validate:"oneof=asc desc"`
	Tags   []string `json:"tags"   validate:"max=5"`
	Query  string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

// Parse extracts and validates pagination, sorting, tags, search queries, and time filters from HTTP request query parameters.
func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	var err error
	if fq, err = parseLimit(qs, fq); err != nil {
		return PaginatedFeedQuery{}, err
	}

	if fq, err = parseOffset(qs, fq); err != nil {
		return PaginatedFeedQuery{}, err
	}

	if fq, err = parseSort(qs, fq); err != nil {
		return PaginatedFeedQuery{}, err
	}

	parseOptionalFields(qs, &fq)

	return fq, nil
}

func parseLimit(qs url.Values, fq PaginatedFeedQuery) (PaginatedFeedQuery, error) {
	limit := qs.Get("limit")
	if limit == "" {
		return fq, nil
	}

	l, err := strconv.Atoi(limit)
	if err != nil {
		return PaginatedFeedQuery{}, errors.Join(errInvalidLimitValue, err)
	}

	fq.Limit = l

	return fq, nil
}

func parseOffset(qs url.Values, fq PaginatedFeedQuery) (PaginatedFeedQuery, error) {
	offset := qs.Get("offset")
	if offset == "" {
		return fq, nil
	}

	o, err := strconv.Atoi(offset)
	if err != nil {
		return PaginatedFeedQuery{}, errors.Join(errInvalidOffsetValue, err)
	}

	fq.Offset = o

	return fq, nil
}

func parseSort(qs url.Values, fq PaginatedFeedQuery) (PaginatedFeedQuery, error) {
	sort := qs.Get("sort")
	if sort == "" {
		return fq, nil
	}

	if sort != SortModeASC && sort != SortModeDESC {
		return PaginatedFeedQuery{}, fmt.Errorf("%w: %v", errInvalidSortValue, sort)
	}

	fq.Sort = sort

	return fq, nil
}

func parseOptionalFields(qs url.Values, fq *PaginatedFeedQuery) {
	if tags := qs.Get("tags"); tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	if query := qs.Get("query"); query != "" {
		fq.Query = query
	}

	if since := qs.Get("since"); since != "" {
		fq.Since = parseTime(since)
	}

	if until := qs.Get("until"); until != "" {
		fq.Until = parseTime(until)
	}
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
