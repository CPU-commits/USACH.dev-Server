package forms

type ProfileForm struct {
	Description string `form:"description" binding:"max=500"`
}
