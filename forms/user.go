package forms

type UserForm struct {
	Email    string `json:"email" binding:"required,max=100,email"`
	Password string `json:"password" binding:"required,min=8,max=50"`
	FullName string `json:"full_name" binding:"required,max=100"`
}
