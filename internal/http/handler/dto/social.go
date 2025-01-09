package dto

type SocialBind struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}
