package api

import "strconv"

type ListRequest[TQuery any] struct {
	Path    string
	Query   TQuery
	Page    PageRequest
	Headers []Header
}

type ListResponse[TItem any] struct {
	Items []TItem      `json:"items"`
	Meta  ResponseMeta `json:"meta"`
}

type PageDecoder[TQuery any, TItem any] interface {
	QueryParams(query TQuery, page PageRequest) []QueryParam
	DecodePage(value TItem) TItem
}

func PageParams(page PageRequest) []QueryParam {
	params := make([]QueryParam, 0, 2)
	if page.Page > 0 {
		params = append(params, QueryParam{Name: "page", Value: strconv.Itoa(page.Page)})
	}
	if page.PerPage > 0 {
		params = append(params, QueryParam{Name: "perPage", Value: strconv.Itoa(page.PerPage)})
	}
	return params
}
