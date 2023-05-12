package forms

type DiscussionForm struct {
	Title      string   `form:"title" binding:"required,max=100"`
	Repository string   `form:"repository" binding:"isMongoId"`
	Text       string   `form:"text" binding:"required"`
	Tags       []string `form:"tags" binding:"dive,max=100"`
}
