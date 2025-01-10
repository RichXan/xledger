package dto

type Page struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// 用于泛型
type (
	TDto        any
	TKey        any
	TUpdate     any
	TQueryParam any
	TQueryOpt   any
)
