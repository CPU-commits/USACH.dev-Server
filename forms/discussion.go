package forms

type DiscussionForm struct {
	Title      string   `form:"title" binding:"required,max=100"`
	Repository string   `form:"repository,omitempty" binding:"omitempty,isMongoId"`
	Snippet    string   `form:"snippet" binding:"required,max=300"`
	Text       string   `form:"text" binding:"required"`
	Tags       []string `form:"tags" binding:"dive,max=100"`
}
