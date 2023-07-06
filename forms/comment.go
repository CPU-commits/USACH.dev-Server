package forms

type CommentForm struct {
	Comment string `json:"comment" binding:"required"`
}
