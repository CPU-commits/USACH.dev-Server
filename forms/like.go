package forms

type LikeForm struct {
	Plus *bool `json:"plus" binding:"required"`
}
