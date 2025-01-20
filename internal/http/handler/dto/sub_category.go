package dto

import "github.com/RichXan/xcommon/xhttp"

type SubCategoryCreate struct {
	Name string `json:"name" binding:"required"`
}

type SubCategoryUpdate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	Status   int    `json:"status" validate:"omitempty,oneof=1 2"`
}

type SubCategoryList struct {
	xhttp.PageReq
}
