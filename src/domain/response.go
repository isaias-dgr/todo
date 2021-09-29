package domain

import (
	"log"
	"net/url"
	"strconv"
)

type Response struct {
	Data     interface{} `json:"data,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func NewResponse(data interface{}, total int, filter *Filter, msg string) Response {
	resp := Response{
		Data: data,
	}
	if filter != nil {
		resp.Metadata = NewMetadata(total, filter, msg)
	}
	return resp
}

type Metadata struct {
	Offset  int    `json:"offset,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Total   int    `json:"total,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewMetadata(total int, filter *Filter, msg string) *Metadata {
	return &Metadata{
		Offset:  filter.Offset,
		Limit:   filter.Limit,
		Total:   total,
		Message: msg,
	}
}

type Filter struct {
	Offset int
	Limit  int
	SortBy string
}

func NewFilter(qs url.Values) *Filter {
	return &Filter{
		Offset: GetIntDefault(qs, "offset", 0),
		Limit:  GetIntDefault(qs, "limit", 10),
		SortBy: GetDefault(qs, "sort_by", ""),
	}
}

func GetIntDefault(qs url.Values, k string, v int) int {
	val, err := strconv.Atoi(GetDefault(qs, k, strconv.Itoa(v)))
	if err != nil {
		log.Printf("> error %s", err)
	}
	return val
}

func GetDefault(qs url.Values, k string, v string) string {
	val := qs.Get(k)
	if val == "" {
		return v
	}
	return val
}
